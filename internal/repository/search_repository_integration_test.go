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

// 統合テストの前提: make seed と make import-data（rosters）で実データが投入済みであること。
func TestSearchRepository(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewSearchRepository(pool)

	t.Run("ListSeasons returns rosters data distinct seasons descending", func(t *testing.T) {
		seasons, err := repo.ListSeasons(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, seasons)

		assert.Contains(t, seasons, int32(2024))
		for i := 1; i < len(seasons); i++ {
			assert.Greater(t, seasons[i-1], seasons[i], "降順で並んでいること")
		}
	})

	t.Run("ListTeams returns all 32 teams ordered by id", func(t *testing.T) {
		teams, err := repo.ListTeams(ctx)
		require.NoError(t, err)
		require.Len(t, teams, 32)

		assert.EqualValues(t, 1, teams[0].ID)
		assert.Equal(t, "Buffalo", teams[0].City)
		assert.Equal(t, "Bills", teams[0].Name)
		assert.EqualValues(t, 1, teams[0].Conference)
		assert.EqualValues(t, 1, teams[0].Area)
		require.NotNil(t, teams[0].ImageFile)
		assert.Equal(t, "bills.gif", *teams[0].ImageFile)

		for i := 1; i < len(teams); i++ {
			assert.Greater(t, teams[i].ID, teams[i-1].ID, "id 昇順で並んでいること")
		}
	})
}
