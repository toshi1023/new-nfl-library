//go:build integration

package rosterimport

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
)

func mustPtr(v float64) *float64 { return &v }

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

	_, err = pool.Exec(ctx, `TRUNCATE rosters, players RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	entry := domain.RosterEntry{
		Season: 2024, TeamID: 1, Team: "buf", Number: "17", Position: "QB",
		FirstName: "Josh", LastName: "Allen", Birthday: "19960521",
		WeightKg: mustPtr(106.7), HeightCm: mustPtr(195.6), College: "Wyoming", Experience: "6",
		DraftedTeam: "Buffalo Bills", DraftedRound: "1st", DraftedRank: "7th pick", DraftedYear: "2018",
	}

	im := New(pool)

	// 1 回目: player 新規作成 + roster 新規作成
	res, err := im.Run(ctx, []domain.RosterEntry{entry}, Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{PlayersCreated: 1, RostersUpserted: 1}, res)

	var playerCount, rosterCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM players`).Scan(&playerCount))
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM rosters`).Scan(&rosterCount))
	assert.Equal(t, 1, playerCount)
	assert.Equal(t, 1, rosterCount)

	// 2 回目（背番号・大学・体重を変更）: player は行を増やさず更新され、roster も更新される。
	entry.Number = "9"
	entry.College = "Updated University"
	entry.WeightKg = mustPtr(110.0)
	res, err = im.Run(ctx, []domain.RosterEntry{entry}, Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{PlayersReused: 1, RostersUpserted: 1}, res)

	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM players`).Scan(&playerCount))
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM rosters`).Scan(&rosterCount))
	assert.Equal(t, 1, playerCount)
	assert.Equal(t, 1, rosterCount)

	var number int
	require.NoError(t, pool.QueryRow(ctx, `SELECT number FROM rosters LIMIT 1`).Scan(&number))
	assert.Equal(t, 9, number)

	var college string
	var weight float64
	require.NoError(t, pool.QueryRow(ctx, `SELECT college, weight FROM players LIMIT 1`).Scan(&college, &weight))
	assert.Equal(t, "Updated University", college, "既存 player の大学が CSV の新しい値で更新される")
	assert.InDelta(t, 110.0, weight, 0.01, "既存 player の体重が CSV の新しい値で更新される")
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

	_, err = pool.Exec(ctx, `TRUNCATE rosters, players RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	entry := domain.RosterEntry{
		Season: 2024, TeamID: 1, Team: "buf", Position: "___UNKNOWN___",
		FirstName: "No", LastName: "Body", Birthday: "20000101",
	}

	im := New(pool)

	t.Run("default: skip and continue", func(t *testing.T) {
		res, err := im.Run(ctx, []domain.RosterEntry{entry}, Options{})
		require.NoError(t, err)
		assert.Equal(t, Result{Skipped: 1}, res)
	})

	t.Run("strict: abort", func(t *testing.T) {
		_, err := im.Run(ctx, []domain.RosterEntry{entry}, Options{Strict: true})
		require.Error(t, err)
	})
}
