package maddenimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// MinSeason はこのコマンドで扱える最古のシーズン。
const MinSeason = 2012

// Options は Importer.Run のオプション。
type Options struct {
	// Strict が true の場合、positions マスタに無いポジションが出た時点で処理全体を中断する。
	// false（既定）ならその行を警告付きでスキップし、続行する。
	Strict bool
}

// Result は Run の結果件数。
type Result struct {
	Updated int
	Skipped int
}

// File は 1 つの CSV ファイルの読み込み結果。
type File struct {
	// Path はログ出力用のファイルパス。
	Path string
	// Rows は ParseCSV の結果。
	Rows []Row
	// FallbackTeamID は Rows の各行に Team 列が無い場合に使う team_id。
	// TeamHintFromFilename 等で解決できなかった場合は 0（未設定）。
	FallbackTeamID int32
}

// Importer は Madden CSV の内容を 1 トランザクションで rosters に反映する。
type Importer struct {
	pool *pgxpool.Pool
}

// New は Importer を生成する。
func New(pool *pgxpool.Pool) *Importer {
	return &Importer{pool: pool}
}

// Run は files の全行を season に対して 1 トランザクションで反映する。
// season + team_id + firstname + lastname に一致する roster が無い行はスキップする
// （このコマンドの前に cmd/importrosters で roster が作成済みであることが前提）。
// 途中で回復不能なエラーが起きた場合は何も反映されない。
func (im *Importer) Run(ctx context.Context, season int, files []File, opts Options) (Result, error) {
	if season < MinSeason {
		return Result{}, fmt.Errorf("season must be %d or later, got %d", MinSeason, season)
	}
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
	// CSV の Team 列の値（"Bills" や "Buffalo Bills"）→ team_id。1シーズンあたり高々32種類しか無いので、
	// 行ごとに DB へ問い合わせない。
	teamIDCache := make(map[string]int32)

	var result Result
	for _, file := range files {
		for i, row := range file.Rows {
			teamID, err := resolveTeamID(ctx, q, teamIDCache, row.Team, file.FallbackTeamID)
			if err != nil {
				return Result{}, fmt.Errorf("%s row %d: %w", file.Path, i+1, err)
			}

			positionID, ok := positionIDs[row.Position]
			if !ok {
				rowErr := fmt.Errorf("%s row %d (%s %s): position %q is not defined in positions master",
					file.Path, i+1, row.FirstName, row.LastName, row.Position)
				if opts.Strict {
					return Result{}, rowErr
				}
				logger.Warn("skip row: unknown position", slog.String("err", rowErr.Error()))
				result.Skipped++
				continue
			}

			roster, err := findRoster(ctx, q, season, teamID, row.FirstName, row.LastName)
			if errors.Is(err, pgx.ErrNoRows) {
				logger.Debug("skip row: no matching roster",
					slog.String("path", file.Path), slog.Int("row", i+1),
					slog.String("firstname", row.FirstName), slog.String("lastname", row.LastName),
					slog.Int("team_id", int(teamID)),
				)
				result.Skipped++
				continue
			}
			if err != nil {
				return Result{}, fmt.Errorf("%s row %d: get roster: %w", file.Path, i+1, err)
			}

			number := roster.Number // CSV に背番号列が無ければ既存値を保持する
			if n, perr := parseOptionalInt32(row.Number); perr != nil {
				return Result{}, fmt.Errorf("%s row %d: number %q: %w", file.Path, i+1, row.Number, perr)
			} else if n != nil {
				number = n
			}
			rating, err := parseOptionalInt32(row.Rating)
			if err != nil {
				return Result{}, fmt.Errorf("%s row %d: rating %q: %w", file.Path, i+1, row.Rating, err)
			}
			attrs, err := json.Marshal(row.Extra)
			if err != nil {
				return Result{}, fmt.Errorf("%s row %d: marshal extra attributes: %w", file.Path, i+1, err)
			}

			if err := q.UpdateRosterMaddenAttributes(ctx, sqlc.UpdateRosterMaddenAttributesParams{
				PositionID:       positionID,
				Rating:           rating,
				Number:           number,
				MaddenAttributes: attrs,
				ID:               roster.ID,
			}); err != nil {
				return Result{}, fmt.Errorf("%s row %d: update roster: %w", file.Path, i+1, err)
			}
			result.Updated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}

// findRoster は season + team_id + firstname + lastname で roster を検索する。
// PFR（players テーブルの元データ）はイニシャル名（"A.J." 等）のピリオド有無が選手ごとに不統一
// （例: "A.J. Brown" はピリオド有り、"AJ Terrell" は無し）なので、そのままの表記で見つからなければ
// ピリオドを除いた形でも 1 回だけ試す。
func findRoster(ctx context.Context, q *sqlc.Queries, season int, teamID int32, firstName, lastName string) (*sqlc.Roster, error) {
	params := sqlc.GetRosterByPlayerIdentityParams{Season: int32(season), TeamID: teamID, Firstname: firstName, Lastname: lastName}
	roster, err := q.GetRosterByPlayerIdentity(ctx, params)
	if err == nil {
		return roster, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	stripped := strings.ReplaceAll(firstName, ".", "")
	if stripped == firstName {
		return nil, pgx.ErrNoRows
	}
	params.Firstname = stripped
	roster, err = q.GetRosterByPlayerIdentity(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return roster, nil
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

// teamNameAliases は Madden CSV の Team 列に現れる表記のうち、teams.name にも city+' '+name にも
// 一致しないものを現在のチーム名（teams.name）に対応付ける。実データ（2012〜2025年の全CSV）を走査して
// 洗い出したもので、略記の都市名（"NY Jets"）と、移転・改名前の旧名称（"Oakland Raiders" 等）が該当する。
// このテーブルは domain パッケージではなく、Madden CSV 固有の表記ゆれとしてここに閉じている。
var teamNameAliases = map[string]string{
	"NY Jets":             "Jets",
	"NY Giants":           "Giants",
	"Oakland Raiders":     "Raiders",    // 2019年以前。2020年に Las Vegas Raiders へ移転
	"Washington Redskins": "Commanders", // 2019年以前の旧名称
	"Redskins":            "Commanders", // 同上（都市名無しの表記）
	"Football Team":       "Commanders", // 2020・2021年の暫定名称（Redskins → Commanders の間）
}

// resolveTeamID は行の team_id を決める。
// csvTeam が空（CSV に Team 列が無い）なら fallbackTeamID を使い、それも 0 ならエラーにする。
// csvTeam があれば teamNameAliases で正規化したうえで、teams.name または city+' '+name に一致する
// チームを検索する（結果は元の csvTeam をキーにキャッシュする）。
func resolveTeamID(ctx context.Context, q *sqlc.Queries, cache map[string]int32, csvTeam string, fallbackTeamID int32) (int32, error) {
	csvTeam = strings.TrimSpace(csvTeam)
	if csvTeam == "" {
		if fallbackTeamID == 0 {
			return 0, fmt.Errorf("no team information: CSV has no Team column and the team could not be determined from the filename")
		}
		return fallbackTeamID, nil
	}
	if id, ok := cache[csvTeam]; ok {
		return id, nil
	}
	lookup := csvTeam
	if alias, ok := teamNameAliases[csvTeam]; ok {
		lookup = alias
	}
	t, err := q.FindTeamByName(ctx, lookup)
	if err != nil {
		return 0, fmt.Errorf("find team %q: %w", csvTeam, err)
	}
	cache[csvTeam] = t.ID
	return t.ID, nil
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
