//go:build integration

package maddenimport

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

func mustFloat(v float64) *float64 { return &v }

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

	// 前提となる roster を rosterimport 経由で用意する（このコマンドは既存 roster しか更新しないため）。
	entries := []domain.RosterEntry{
		{Season: 2024, TeamID: 1, Team: "buf", Number: "17", Position: "QB",
			FirstName: "Josh", LastName: "Allen", Birthday: "19960521",
			WeightKg: mustFloat(106.7), HeightCm: mustFloat(195.6)},
		// PFR 側でピリオド無しの表記になっている選手（実データで確認済みのパターン）。
		{Season: 2024, TeamID: 27, Team: "atl", Number: "1", Position: "CB",
			FirstName: "AJ", LastName: "Terrell", Birthday: "19990101"},
	}
	_, err = rosterimport.New(pool).Run(ctx, entries, rosterimport.Options{})
	require.NoError(t, err)

	files := []File{{
		Path: "test.csv",
		Rows: []Row{
			{Team: "Bills", Position: "QB", FirstName: "Josh", LastName: "Allen", Number: "17", Rating: "92",
				Extra: map[string]string{"Speed": "88"}},
			// Madden 側はピリオド有り表記。findRoster のフォールバックで一致するはず。
			{Team: "Atlanta Falcons", Position: "CB", FirstName: "A.J.", LastName: "Terrell", Rating: "85",
				Extra: map[string]string{"College": "Clemson"}},
			// roster に存在しない選手はスキップされる。
			{Team: "Bills", Position: "QB", FirstName: "Nobody", LastName: "Here", Rating: "50"},
		},
	}}

	im := New(pool)
	res, err := im.Run(ctx, 2024, files, Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Updated: 2, Skipped: 1}, res)

	q := sqlc.New(pool)
	allen, err := q.GetRosterByPlayerIdentity(ctx, sqlc.GetRosterByPlayerIdentityParams{
		Season: 2024, TeamID: 1, Firstname: "Josh", Lastname: "Allen",
	})
	require.NoError(t, err)
	require.NotNil(t, allen.Rating)
	assert.EqualValues(t, 92, *allen.Rating)
	assert.JSONEq(t, `{"Speed":"88"}`, string(allen.MaddenAttributes))

	terrell, err := q.GetRosterByPlayerIdentity(ctx, sqlc.GetRosterByPlayerIdentityParams{
		Season: 2024, TeamID: 27, Firstname: "AJ", Lastname: "Terrell",
	})
	require.NoError(t, err, "ピリオド有り(A.J.)で検索しても、ピリオド無し(AJ)で保存された roster に一致する")
	require.NotNil(t, terrell.Rating)
	assert.EqualValues(t, 85, *terrell.Rating)
	require.NotNil(t, terrell.Number, "CSVにJersey列が無いので既存値(1)を保持する")
	assert.EqualValues(t, 1, *terrell.Number)
}

func TestImporter_Run_TeamAliasAndFilenameFallback(t *testing.T) {
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

	entries := []domain.RosterEntry{
		{Season: 2019, TeamID: 19, Team: "was", Position: "QB", FirstName: "Case", LastName: "Keenum", Birthday: "19881015"},
		{Season: 2012, TeamID: 19, Team: "was", Position: "QB", FirstName: "Rex", LastName: "Grossman", Birthday: "19800808"},
		{Season: 2012, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Ryan", LastName: "Fitzpatrick", Birthday: "19821115"},
	}
	_, err = rosterimport.New(pool).Run(ctx, entries, rosterimport.Options{})
	require.NoError(t, err)

	files := []File{
		{
			Path: "washington_2019.csv",
			Rows: []Row{{Team: "Washington Redskins", Position: "QB", FirstName: "Case", LastName: "Keenum", Rating: "70"}},
		},
		{
			Path: "was_madden_nfl_13.csv",
			Rows: []Row{{Team: "Redskins", Position: "QB", FirstName: "Rex", LastName: "Grossman", Rating: "72"}},
		},
		{
			// Team 列を持たない年（2014・2015年分と同じ）を模したファイル。ファイル名から buf を推定する。
			Path:           "buf_madden_nfl_13.csv",
			Rows:           []Row{{Position: "QB", FirstName: "Ryan", LastName: "Fitzpatrick", Rating: "75"}},
			FallbackTeamID: 1,
		},
	}

	res, err := New(pool).Run(ctx, 2012, files[2:], Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Updated: 1, Skipped: 0}, res, "ファイル名からのチーム推定")

	res, err = New(pool).Run(ctx, 2012, files[1:2], Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Updated: 1, Skipped: 0}, res, "都市名無しの旧チーム名 Redskins のエイリアス解決")

	res, err = New(pool).Run(ctx, 2019, files[:1], Options{})
	require.NoError(t, err)
	assert.Equal(t, Result{Updated: 1, Skipped: 0}, res, "旧チーム名 Washington Redskins のエイリアス解決")
}
