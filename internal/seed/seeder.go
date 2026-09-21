// Package seed はマスタデータの初期投入（Laravel の database/seeders 相当）を行う。
package seed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// root ユーザーの固定値。Laravel の UsersTableSeeder と同じ。
const (
	rootName           = "root"
	rootFavoriteSeason = int32(2013)
	rootFavoriteTeam   = int32(31) // 49ers
)

// rootSettings は root ユーザーの初期設定。Laravel の SettingsTableSeeder と同じ。
var rootSettings = []struct {
	SettingID int32
	Flg       bool
}{
	{SettingID: 1, Flg: true},  // Season 検索の表示有無
	{SettingID: 2, Flg: false}, // Team 検索の表示有無
}

// Options はシード実行のオプション。
type Options struct {
	// Force が true なら対象テーブルを TRUNCATE してから投入し直す。false なら既にデータがあるテーブルはスキップする。
	Force bool
	// RootEmail / RootPassword は root ユーザーの認証情報。どちらかが空なら users と settings の投入をスキップする。
	RootEmail    string
	RootPassword string
}

// Seeder はマスタデータを投入する。
type Seeder struct {
	pool *pgxpool.Pool
}

// New は Seeder を生成する。
func New(pool *pgxpool.Pool) *Seeder {
	return &Seeder{pool: pool}
}

// Run は 1 つのトランザクションで全テーブルのシードを実行する。途中で失敗した場合は何も反映されない。
// 投入順は teams → positions → formations → pf_relations → users → settings → foul_rules（外部キーの依存順）。
func (s *Seeder) Run(ctx context.Context, opts Options) error {
	logger := logging.FromContext(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rerr := tx.Rollback(ctx); rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
			logger.Warn("rollback failed", slog.Any("err", rerr))
		}
	}()
	q := sqlc.New(tx)

	if opts.Force {
		logger.Warn("truncating seed tables (force)")
		if err := q.TruncateSeedTables(ctx); err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
	}

	if err := seedTeams(ctx, q); err != nil {
		return err
	}
	if err := seedPositions(ctx, q); err != nil {
		return err
	}
	formationIDs, err := seedFormations(ctx, q)
	if err != nil {
		return err
	}
	if err := seedPfRelations(ctx, q, formationIDs); err != nil {
		return err
	}
	userID, err := seedUsers(ctx, q, opts)
	if err != nil {
		return err
	}
	if err := seedSettings(ctx, q, userID); err != nil {
		return err
	}
	if err := seedFoulRules(ctx, q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func seedTeams(ctx context.Context, q *sqlc.Queries) error {
	if skip, err := shouldSkip(ctx, "teams", q.TeamsExist); skip || err != nil {
		return err
	}
	n, err := q.InsertTeams(ctx, teams)
	if err != nil {
		return fmt.Errorf("insert teams: %w", err)
	}
	logInserted(ctx, "teams", n)
	return nil
}

func seedPositions(ctx context.Context, q *sqlc.Queries) error {
	if skip, err := shouldSkip(ctx, "positions", q.PositionsExist); skip || err != nil {
		return err
	}
	n, err := q.InsertPositions(ctx, positions)
	if err != nil {
		return fmt.Errorf("insert positions: %w", err)
	}
	logInserted(ctx, "positions", n)
	return nil
}

// seedFormations は formations を投入し、Key → id の対応を返す（pf_relations の投入に使う）。
// スキップした場合も既存行から対応を組み立てる。
func seedFormations(ctx context.Context, q *sqlc.Queries) (map[string]int32, error) {
	skip, err := shouldSkip(ctx, "formations", q.FormationsExist)
	if err != nil {
		return nil, err
	}
	if !skip {
		params := make([]sqlc.InsertFormationsParams, 0, len(formations))
		for _, f := range formations {
			params = append(params, sqlc.InsertFormationsParams{Odflg: f.Odflg, Name: f.Name})
		}
		n, err := q.InsertFormations(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("insert formations: %w", err)
		}
		logInserted(ctx, "formations", n)
	}

	rows, err := q.ListFormations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list formations: %w", err)
	}
	if len(rows) != len(formations) {
		return nil, fmt.Errorf("formations table has %d rows but seed data has %d; run with force to reseed", len(rows), len(formations))
	}
	ids := make(map[string]int32, len(formations))
	for i, f := range formations {
		if rows[i].Name != f.Name {
			return nil, fmt.Errorf("formations row %d is %q but seed data expects %q; run with force to reseed", i+1, rows[i].Name, f.Name)
		}
		ids[f.Key] = rows[i].ID
	}
	return ids, nil
}

func seedPfRelations(ctx context.Context, q *sqlc.Queries, formationIDs map[string]int32) error {
	if skip, err := shouldSkip(ctx, "pf_relations", q.PfRelationsExist); skip || err != nil {
		return err
	}
	params := make([]sqlc.InsertPfRelationsParams, 0, len(pfRelations))
	for _, r := range pfRelations {
		id, ok := formationIDs[r.Formation]
		if !ok {
			return fmt.Errorf("pf_relations: unknown formation key %q", r.Formation)
		}
		params = append(params, sqlc.InsertPfRelationsParams{
			FormationID:      id,
			PositionCategory: r.PositionCategory,
			AbstractCategory: r.AbstractCategory,
		})
	}
	n, err := q.InsertPfRelations(ctx, params)
	if err != nil {
		return fmt.Errorf("insert pf_relations: %w", err)
	}
	logInserted(ctx, "pf_relations", n)
	return nil
}

// seedUsers は root ユーザーを作成し、その id を返す。スキップした場合は既存の root を探して id を返す。
// 認証情報が未指定なら 0 を返す（settings もスキップされる）。
func seedUsers(ctx context.Context, q *sqlc.Queries, opts Options) (int32, error) {
	logger := logging.FromContext(ctx)
	if opts.RootEmail == "" || opts.RootPassword == "" {
		logger.Warn("skip users: SEED_ROOT_EMAIL / SEED_ROOT_PASSWORD not set")
		return 0, nil
	}

	skip, err := shouldSkip(ctx, "users", q.UsersExist)
	if err != nil {
		return 0, err
	}
	if skip {
		u, err := q.GetUserByEmail(ctx, opts.RootEmail)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("users table is not empty and root user was not found; skip settings", slog.String("email", opts.RootEmail))
			return 0, nil
		}
		if err != nil {
			return 0, fmt.Errorf("get root user: %w", err)
		}
		return u.ID, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(opts.RootPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	season, team := rootFavoriteSeason, rootFavoriteTeam
	u, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		Name:           rootName,
		Email:          opts.RootEmail,
		Password:       string(hash),
		FavoriteSeason: &season,
		FavoriteTeam:   &team,
	})
	if err != nil {
		return 0, fmt.Errorf("insert root user: %w", err)
	}
	logInserted(ctx, "users", 1)
	return u.ID, nil
}

func seedSettings(ctx context.Context, q *sqlc.Queries, userID int32) error {
	if userID == 0 {
		return nil
	}
	if skip, err := shouldSkip(ctx, "settings", q.SettingsExist); skip || err != nil {
		return err
	}
	params := make([]sqlc.InsertSettingsParams, 0, len(rootSettings))
	for _, s := range rootSettings {
		params = append(params, sqlc.InsertSettingsParams{UserID: userID, SettingID: s.SettingID, Flg: s.Flg})
	}
	n, err := q.InsertSettings(ctx, params)
	if err != nil {
		return fmt.Errorf("insert settings: %w", err)
	}
	logInserted(ctx, "settings", n)
	return nil
}

func seedFoulRules(ctx context.Context, q *sqlc.Queries) error {
	if skip, err := shouldSkip(ctx, "foul_rules", q.FoulRulesExist); skip || err != nil {
		return err
	}
	n, err := q.InsertFoulRules(ctx, foulRules)
	if err != nil {
		return fmt.Errorf("insert foul_rules: %w", err)
	}
	logInserted(ctx, "foul_rules", n)
	return nil
}

// shouldSkip はテーブルに既にデータがあればログを出して true を返す。
func shouldSkip(ctx context.Context, table string, exists func(context.Context) (bool, error)) (bool, error) {
	ok, err := exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check %s: %w", table, err)
	}
	if ok {
		logging.FromContext(ctx).Info("skip: table already has rows", slog.String("table", table))
	}
	return ok, nil
}

func logInserted(ctx context.Context, table string, n int64) {
	logging.FromContext(ctx).Info("seeded", slog.String("table", table), slog.Int64("rows", n))
}
