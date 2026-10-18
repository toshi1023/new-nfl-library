//go:build integration

package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
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
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, nil, "rating", "desc")
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
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2099, 1, nil, "rating", "desc")
		require.NoError(t, err)
		assert.Empty(t, rosters)
	})

	t.Run("single name filters by partial, case-insensitive match on firstname or lastname", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, []string{"allen"}, "rating", "desc")
		require.NoError(t, err)
		require.NotEmpty(t, rosters)
		for _, r := range rosters {
			assert.True(t,
				containsFold(r.Player.FirstName, "allen") || containsFold(r.Player.LastName, "allen"),
				"name=allen で firstname/lastname のどちらにも一致しない選手が混在している: %+v", r.Player,
			)
		}

		found := false
		for _, r := range rosters {
			if r.Player.FirstName == "Josh" && r.Player.LastName == "Allen" {
				found = true
			}
		}
		assert.True(t, found, "Josh Allen が見つかること")
	})

	t.Run("multiple names are OR'd together (union of matches)", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, []string{"Josh", "Kincaid"}, "rating", "desc")
		require.NoError(t, err)

		foundAllen, foundKincaid := false, false
		for _, r := range rosters {
			assert.True(t,
				containsFold(r.Player.FirstName, "Josh") || containsFold(r.Player.LastName, "Josh") ||
					containsFold(r.Player.FirstName, "Kincaid") || containsFold(r.Player.LastName, "Kincaid"),
				"OR条件のいずれにも一致しない選手が混在している: %+v", r.Player,
			)
			if r.Player.FirstName == "Josh" && r.Player.LastName == "Allen" {
				foundAllen = true
			}
			if r.Player.LastName == "Kincaid" {
				foundKincaid = true
			}
		}
		assert.True(t, foundAllen, "Josh Allen が見つかること（1語目にマッチ）")
		assert.True(t, foundKincaid, "Kincaid が見つかること（2語目にマッチ）")
	})

	t.Run("name with no match returns empty, not error", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, []string{"ZZZ_NO_SUCH_PLAYER_ZZZ"}, "rating", "desc")
		require.NoError(t, err)
		assert.Empty(t, rosters)
	})

	t.Run("sort_by=rating sort_order=desc is descending with nulls last", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, nil, "rating", "desc")
		require.NoError(t, err)
		require.NotEmpty(t, rosters)
		assertSortedByRating(t, rosters, true)
	})

	t.Run("sort_by=rating sort_order=asc is ascending with nulls last", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, nil, "rating", "asc")
		require.NoError(t, err)
		require.NotEmpty(t, rosters)
		assertSortedByRating(t, rosters, false)
	})

	t.Run("sort_by=name sort_order=asc is alphabetical by lastname then firstname", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, nil, "name", "asc")
		require.NoError(t, err)
		require.NotEmpty(t, rosters)
		for i := 1; i < len(rosters); i++ {
			prev, cur := rosters[i-1].Player, rosters[i].Player
			ok := prev.LastName < cur.LastName || (prev.LastName == cur.LastName && prev.FirstName <= cur.FirstName)
			assert.True(t, ok, "lastname, firstname の昇順になっていない: %+v -> %+v", prev, cur)
		}
	})

	t.Run("sort_by=number sort_order=desc puts missing numbers last", func(t *testing.T) {
		rosters, err := repo.ListBySeasonAndTeam(ctx, 2024, 1, nil, "number", "desc")
		require.NoError(t, err)
		require.NotEmpty(t, rosters)

		sawNull := false
		for i, r := range rosters {
			if r.Number == nil {
				sawNull = true
				continue
			}
			require.False(t, sawNull, "number が NULL の行より後に NULL でない行が来ている（NULLS LAST になっていない）: index=%d", i)
		}
	})
}

// assertSortedByRating は rating が（NULL を除いて）指定方向に単調であることと、
// NULL（rating 無し）が常に末尾に来ることを検証する。
func assertSortedByRating(t *testing.T, rosters []domain.RosterDetail, desc bool) {
	t.Helper()
	sawNull := false
	for i, r := range rosters {
		if r.Rating == nil {
			sawNull = true
			continue
		}
		require.False(t, sawNull, "rating が NULL の行より後に NULL でない行が来ている（NULLS LAST になっていない）: index=%d", i)
		if i == 0 {
			continue
		}
		prev := rosters[i-1].Rating
		if prev == nil {
			continue
		}
		if desc {
			assert.GreaterOrEqual(t, *prev, *r.Rating)
		} else {
			assert.LessOrEqual(t, *prev, *r.Rating)
		}
	}
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
