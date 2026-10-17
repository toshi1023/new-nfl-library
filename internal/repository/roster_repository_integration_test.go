//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
)

// 統合テストの前提: make seed と make import-data（rosters/madden）で実データが投入済みであること。
func TestRosterRepository_ListBySeasonAndTeam(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewRosterRepository(pool)

	t.Run("buf 2024 includes Josh Allen with madden attributes", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1)
		require.NoError(t, err)
		require.NotEmpty(t, rosters)

		found := false
		for _, r := range rosters {
			if r.Player.FirstName == "Josh" && r.Player.LastName == "Allen" {
				found = true
				assert.EqualValues(t, 2024, r.Season)
				assert.EqualValues(t, 1, r.TeamID)
				assert.Equal(t, "QB", r.Position.Name)
				require.NotNil(t, r.Number)
				assert.EqualValues(t, 17, *r.Number)
				require.NotNil(t, r.Player.HeightCm)
				assert.InDelta(t, 195.6, *r.Player.HeightCm, 0.01)
				assert.NotEmpty(t, r.MaddenAttributes, "2024シーズンはMaddenデータ投入済みのはず")
			}
		}
		assert.True(t, found, "Josh Allen が見つかること")
	})

	t.Run("unknown season/team returns empty, not error", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2099, 1)
		require.NoError(t, err)
		assert.Empty(t, rosters)
	})
}
