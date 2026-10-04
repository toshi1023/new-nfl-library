// Package starterimport は cmd/scrapestarters が出力した CSV を読み込み、starters テーブルへ反映する。
// Laravel の ScrapeStarters（DB 書き込み部分のうち starters テーブルへの反映）に相当する。
// ScrapeStarters は roster.position_id もスターター側の値で上書きしていたが、このパッケージは
// starters テーブルのみを更新する（rosters.position_id は cmd/importrosters / cmd/importmadden の
// 2経路が既に更新しており、3つ目の更新経路を増やすと優先順位が分かりにくくなるため）。
package starterimport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// Options は Importer.Run のオプション。
type Options struct {
	// Strict が true の場合、positions マスタに無いポジションが出た時点で処理全体を中断する。
	// false（既定）ならその行を警告付きでスキップし、続行する。
	Strict bool
}

// Result は Run の結果件数。
type Result struct {
	Upserted int
	Skipped  int
}

// Importer は starter CSV の内容を 1 トランザクションで starters に反映する。
type Importer struct {
	pool *pgxpool.Pool
}

// New は Importer を生成する。
func New(pool *pgxpool.Pool) *Importer {
	return &Importer{pool: pool}
}

// Run は entries を 1 トランザクションで反映する。season + team_id + firstname + lastname に一致する
// roster が無い行はスキップする（このコマンドの前に cmd/importrosters で roster が作成済みであることが前提）。
// 途中で回復不能なエラーが起きた場合は何も反映されない。
func (im *Importer) Run(ctx context.Context, entries []domain.StarterEntry, opts Options) (Result, error) {
	logger := logging.FromContext(ctx)

	tx, err := im.pool.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rerr := tx.Rollback(ctx); rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
			logger.Warn("rollback failed", slog.Any("err", rerr))
		}
	}()
	q := sqlc.New(tx)

	positionIDs, err := loadPositionIDs(ctx, q)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for i, e := range entries {
		if _, ok := positionIDs[e.Position]; !ok {
			rowErr := fmt.Errorf("row %d (%s %s, team %d): position %q is not defined in positions master",
				i+1, e.FirstName, e.LastName, e.TeamID, e.Position)
			if opts.Strict {
				return Result{}, rowErr
			}
			logger.Warn("skip row: unknown position", slog.String("err", rowErr.Error()))
			result.Skipped++
			continue
		}

		roster, err := q.GetRosterByPlayerIdentity(ctx, sqlc.GetRosterByPlayerIdentityParams{
			Season:    int32(e.Season),
			TeamID:    int32(e.TeamID),
			Firstname: e.FirstName,
			Lastname:  e.LastName,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("skip row: no matching roster",
				slog.Int("row", i+1), slog.String("firstname", e.FirstName), slog.String("lastname", e.LastName),
				slog.Int("season", e.Season), slog.Int("team_id", e.TeamID),
			)
			result.Skipped++
			continue
		}
		if err != nil {
			return Result{}, fmt.Errorf("row %d: get roster: %w", i+1, err)
		}

		if err := q.UpsertStarter(ctx, sqlc.UpsertStarterParams{
			Season:   int32(e.Season),
			Odflg:    e.Odflg,
			RosterID: roster.ID,
		}); err != nil {
			return Result{}, fmt.Errorf("row %d: upsert starter: %w", i+1, err)
		}
		result.Upserted++
	}

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
