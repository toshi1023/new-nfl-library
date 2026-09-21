//go:build integration

package seed

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

func TestSeeder_Run(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	s := New(pool)
	opts := Options{Force: true, RootEmail: "root@example.com", RootPassword: "password"}

	// force で投入
	require.NoError(t, s.Run(ctx, opts))

	q := sqlc.New(pool)
	rows, err := q.ListFormations(ctx)
	require.NoError(t, err)
	assert.Len(t, rows, len(formations))

	u, err := q.GetUserByEmail(ctx, opts.RootEmail)
	require.NoError(t, err)
	assert.Equal(t, rootName, u.Name)

	var counts struct{ teams, positions, pf, foul, settings int }
	require.NoError(t, pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM teams), (SELECT count(*) FROM positions),
		(SELECT count(*) FROM pf_relations), (SELECT count(*) FROM foul_rules), (SELECT count(*) FROM settings)`).
		Scan(&counts.teams, &counts.positions, &counts.pf, &counts.foul, &counts.settings))
	assert.Equal(t, len(teams), counts.teams)
	assert.Equal(t, len(positions), counts.positions)
	assert.Equal(t, len(pfRelations), counts.pf)
	assert.Equal(t, len(foulRules), counts.foul)
	assert.Equal(t, len(rootSettings), counts.settings)

	// 2 回目（force なし）は全テーブルスキップされ、件数が変わらない
	opts.Force = false
	require.NoError(t, s.Run(ctx, opts))
	var teamsAgain int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM teams`).Scan(&teamsAgain))
	assert.Equal(t, len(teams), teamsAgain)
}
