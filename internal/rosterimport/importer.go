// Package rosterimport は cmd/scraperosters が出力した CSV を読み込み、players / rosters テーブルへ
// 反映する。Laravel の ScrapeRoster（DB 書き込み部分）に相当する。
package rosterimport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// Options は Importer.Run のオプション。
type Options struct {
	// Strict が true の場合、positions マスタに無いポジションが出た時点で処理全体を中断する
	// （Laravel の ScrapeRoster と同じ挙動）。false（既定）ならその行を警告付きでスキップし、続行する。
	Strict bool
}

// Result は Run の結果件数。
type Result struct {
	PlayersCreated  int
	PlayersReused   int
	RostersUpserted int
	Skipped         int
}

// Importer は CSV の内容を 1 トランザクションで players / rosters に反映する。
type Importer struct {
	// pool は DB 接続プールを指す。Run 内でトランザクションを開始する。
	// postgres.NewPool で生成したものを渡す。
	pool *pgxpool.Pool
}

// New は Importer を生成する。
func New(pool *pgxpool.Pool) *Importer {
	return &Importer{pool: pool}
}

// Run は entries を 1 トランザクションで反映する。途中で回復不能なエラーが起きた場合は何も反映されない。
func (im *Importer) Run(ctx context.Context, entries []domain.RosterEntry, opts Options) (Result, error) {
	logger := logging.FromContext(ctx)

	tx, err := im.pool.Begin(ctx)
	// トランザクション開始に失敗した場合は処理を中断する。
	// ここではdeferの登録前に記述しているため、tx.Rollback() は呼ばれない。
	if err != nil {
		return Result{}, fmt.Errorf("begin tx: %w", err)
	}
	// プロセスキルやSIGINT（Ctrl+C）を実行したときにロールバックするようにする。tx.Commit() した場合は ErrTxClosed になるので無視する。
	// これ以降に記述した処理で関数が return や panicで抜けるときも、deferの中の処理が実行される
	defer func() {
		if rerr := tx.Rollback(ctx); rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
			logger.Warn("rollback failed", slog.Any("err", rerr))
		}
	}()
	q := sqlc.New(tx)

	// positions は行数（数千件規模）に対して種類が少ない（数十件）ため、最初に全件読み込んでおき、
	// 行ごとに DB へ問い合わせない。
	positionIDs, err := loadPositionIDs(ctx, q)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for i, e := range entries {
		positionID, ok := positionIDs[e.Position]
		// positions マスタに無いポジションが出た場合は、Strict オプションに応じて処理を中断するかスキップする。
		if !ok {
			err := fmt.Errorf("row %d (%s %s %s, team %s): position %q is not defined in positions master",
				i+1, e.FirstName, e.LastName, e.Birthday, e.Team, e.Position)
			// コマンド実行時に引数で指定する Strict が true の場合は処理全体を中断する。false（既定）ならスキップして続行する。
			if opts.Strict {
				return Result{}, err
			}
			logger.Warn("skip row: unknown position", slog.String("err", err.Error()))
			result.Skipped++
			continue
		}

		// players テーブルに既存の player があるかを検索し、無ければ作成する。
		playerID, created, err := findOrCreatePlayer(ctx, q, e)
		if err != nil {
			return Result{}, fmt.Errorf("row %d (%s %s, team %s): %w", i+1, e.FirstName, e.LastName, e.Team, err)
		}
		if created {
			result.PlayersCreated++
		} else {
			result.PlayersReused++
		}

		number, err := parseOptionalInt32(e.Number)
		if err != nil {
			return Result{}, fmt.Errorf("row %d: number %q: %w", i+1, e.Number, err)
		}
		experience, err := parseOptionalInt32(e.Experience)
		if err != nil {
			return Result{}, fmt.Errorf("row %d: experience %q: %w", i+1, e.Experience, err)
		}

		// rosters テーブルは season+team_id+player_id の組み合わせで一意制約があるため、既存があれば更新、無ければ挿入する。
		if _, err := q.UpsertRoster(ctx, sqlc.UpsertRosterParams{
			Season:     int32(e.Season),
			TeamID:     int32(e.TeamID),
			PlayerID:   playerID,
			PositionID: positionID,
			Number:     number,
			Experience: experience,
		}); err != nil {
			return Result{}, fmt.Errorf("row %d (%s %s, team %s): upsert roster: %w", i+1, e.FirstName, e.LastName, e.Team, err)
		}
		result.RostersUpserted++
	}

	// トランザクションをコミットする。ここでエラーが起きた場合は処理全体を中断してdeferでロールバックする。
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}

// loadPositionIDs は positions テーブル全件を読み、名前 → id のマップを返す。
func loadPositionIDs(ctx context.Context, q *sqlc.Queries) (map[string]int32, error) {
	rows, err := q.ListPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("positions table is empty; run `make seed` first")
	}
	ids := make(map[string]int32, len(rows))
	for _, r := range rows {
		ids[r.Name] = r.ID
	}
	return ids, nil
}

// findOrCreatePlayer は firstname+lastname+birthday で既存 player を検索し、無ければ作成する。
// 既存が見つかった場合は CSV の値で更新しない（Laravel の ScrapeRoster と同じ挙動）。
func findOrCreatePlayer(ctx context.Context, q *sqlc.Queries, e domain.RosterEntry) (id int32, created bool, err error) {
	existing, err := q.GetPlayerByIdentity(ctx, sqlc.GetPlayerByIdentityParams{
		Firstname: e.FirstName,
		Lastname:  e.LastName,
		Birthday:  e.Birthday,
	})
	if err == nil {
		return existing.ID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("get player: %w", err)
	}

	height, err := numericFromFloat(e.HeightCm)
	if err != nil {
		return 0, false, fmt.Errorf("height %v: %w", e.HeightCm, err)
	}
	weight, err := numericFromFloat(e.WeightKg)
	if err != nil {
		return 0, false, fmt.Errorf("weight %v: %w", e.WeightKg, err)
	}
	draftedYear, err := parseOptionalInt32(e.DraftedYear)
	if err != nil {
		return 0, false, fmt.Errorf("drafted_year %q: %w", e.DraftedYear, err)
	}

	created2, err := q.CreatePlayer(ctx, sqlc.CreatePlayerParams{
		Firstname:    e.FirstName,
		Lastname:     e.LastName,
		Birthday:     e.Birthday,
		Height:       height,
		Weight:       weight,
		College:      optionalString(e.College),
		DraftedTeam:  optionalString(e.DraftedTeam),
		DraftedRound: optionalString(e.DraftedRound),
		DraftedRank:  optionalString(e.DraftedRank),
		DraftedYear:  draftedYear,
	})
	if err != nil {
		return 0, false, fmt.Errorf("create player: %w", err)
	}
	return created2.ID, true, nil
}

// numericFromFloat は *float64 を pgtype.Numeric に変換する。nil なら NULL（Valid: false）。
func numericFromFloat(v *float64) (pgtype.Numeric, error) {
	if v == nil {
		return pgtype.Numeric{}, nil
	}
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(*v, 'f', 1, 64)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

// parseOptionalInt32 は空文字を nil、それ以外を *int32 に変換する。
func parseOptionalInt32(s string) (*int32, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil, err
	}
	v32 := int32(v)
	return &v32, nil
}

// optionalString は空文字を nil、それ以外を *string に変換する。
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
