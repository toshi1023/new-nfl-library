//go:build integration

package starterimport

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
	"github.com/toshi1023/new-nfl-library/internal/rosterimport"
)

// 統合テストの前提: `make seed` で teams / positions が投入済みであること。
func TestImporter_Run(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(ctx, `TRUNCATE starters, rosters, players RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	// 前提となる roster を rosterimport 経由で用意する。
	rosterEntries := []domain.RosterEntry{
		{Season: 2024, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Josh", LastName: "Allen", Birthday: "19960521"},
	}
	_, err = rosterimport.New(pool).Run(ctx, rosterEntries, rosterimport.Options{})
	require.NoError(t, err)

	entries := []domain.StarterEntry{
		{Season: 2024, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Josh", LastName: "Allen", Odflg: true},
		// roster が存在しない選手はスキップされる。
		{Season: 2024, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Nobody", LastName: "Here", Odflg: true},
	}

	im := New(pool)

	// 1 回目: 新規作成
	res, err := im.Run(ctx, entries, Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Upserted: 1, Skipped: 1}, res)

	q := sqlc.New(pool)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM starters`).Scan(&count))
	assert.Equal(t, 1, count)

	// 2 回目（odflg を変更）: 行が増えず更新される（roster_id の UNIQUE 制約による upsert）。
	entries[0].Odflg = false
	res, err = im.Run(ctx, entries, Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Upserted: 1, Skipped: 1}, res)

	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM starters`).Scan(&count))
	assert.Equal(t, 1, count, "roster_id の UNIQUE 制約により重複せず更新される")

	roster, err := q.GetRosterByPlayerIdentity(ctx, sqlc.GetRosterByPlayerIdentityParams{
		Season: 2024, TeamID: 1, Firstname: "Josh", Lastname: "Allen",
	})
	require.NoError(t, err)
	var odflg bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT odflg FROM starters WHERE roster_id = $1`, roster.ID).Scan(&odflg))
	assert.False(t, odflg)
}

func TestImporter_Run_UnknownPosition(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(ctx, `TRUNCATE starters, rosters, players RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	entry := domain.StarterEntry{Season: 2024, TeamID: 1, Team: "buf", Position: "___UNKNOWN___", FirstName: "No", LastName: "Body"}
	im := New(pool)

	t.Run("default: skip and continue", func(t *testing.T) {
		res, err := im.Run(ctx, []domain.StarterEntry{entry}, Options{})
		require.NoError(t, err)
		assert.Equal(t, Result{Skipped: 1}, res)
	})

	t.Run("strict: abort", func(t *testing.T) {
		_, err := im.Run(ctx, []domain.StarterEntry{entry}, Options{Strict: true})
		require.Error(t, err)
	})
}
