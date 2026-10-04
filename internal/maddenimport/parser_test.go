package maddenimport

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSV_2023Style(t *testing.T) {
	// Team, Position, Full Name, Overall Rating, Jersey Number, ...（2022/2023年分の実データと同じ形式）
	csv := "\ufeffTeam,Position,Full Name,Overall Rating,Jersey Number,Speed,College\n" +
		"Bills,C,Mitch Morse,79,60,66,Missouri\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, hasTeam)

	r := rows[0]
	assert.Equal(t, "Bills", r.Team)
	assert.Equal(t, "C", r.Position)
	assert.Equal(t, "Mitch", r.FirstName)
	assert.Equal(t, "Morse", r.LastName)
	assert.Equal(t, "79", r.Rating)
	assert.Equal(t, "60", r.Number)
	assert.Equal(t, map[string]string{"Speed": "66", "College": "Missouri"}, r.Extra)
}

func TestParseCSV_2025Style_PositionIDPreferred(t *testing.T) {
	// Position がフルワード、Position ID が短縮コード。Team はフル表記（city+name）。
	csv := "ID,First Name,Last Name,Position,Position ID,Team,Team ID,Overall,SPEED\n" +
		"1,Josh,Allen,Quarterback,QB,Buffalo Bills,3,91,91\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, hasTeam)

	r := rows[0]
	assert.Equal(t, "Buffalo Bills", r.Team)
	assert.Equal(t, "QB", r.Position, "Position ID を優先する")
	assert.Equal(t, "Josh", r.FirstName)
	assert.Equal(t, "Allen", r.LastName)
	assert.Equal(t, "91", r.Rating)
	assert.Equal(t, map[string]string{
		"ID": "1", "Position": "Quarterback", "Team ID": "3", "SPEED": "91",
	}, r.Extra, "Position ID を使った場合、フルワードの Position はJSONに残す")
}

func TestParseCSV_2022Style_CombinedNamePreferredOverSplit(t *testing.T) {
	// Full Name と First Name/Last Name が同じファイルに両方存在する（2022年分の実データと同じ）。
	csv := "Team,Full Name,Position,First Name,Last Name,Jersey Number,Overall Rating\n" +
		"Cardinals,Kyler Murray,QB,Kyler,Murray,1,84\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, hasTeam)

	r := rows[0]
	assert.Equal(t, "Kyler", r.FirstName)
	assert.Equal(t, "Murray", r.LastName)
	// 結合列を優先して読むので、未使用の First Name/Last Name 列はそのまま Extra に残る
	assert.Equal(t, map[string]string{"First Name": "Kyler", "Last Name": "Murray"}, r.Extra)
}

func TestParseCSV_2014Style_NoTeamColumn(t *testing.T) {
	// FIRST/LAST/JERSEY（全部大文字）、Team 列なし（2014年分の実データと同じ）。
	csv := "\ufeffFIRST,LAST,POSITION,JERSEY,OVERALL RATING,SPEED\n" +
		"Joe,Hawley,C,61,73,61\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.False(t, hasTeam, "Team列自体が無い")

	r := rows[0]
	assert.Empty(t, r.Team, "Team列が無いので空")
	assert.Equal(t, "Joe", r.FirstName)
	assert.Equal(t, "Hawley", r.LastName)
	assert.Equal(t, "61", r.Number)
	assert.Equal(t, "73", r.Rating)
}

func TestParseCSV_JerseyAliasVariants(t *testing.T) {
	for _, header := range []string{"Jersey", "Jersey Number", "JerseyNumber", "JERSEY", "Jersey #", "jerseyNum", "JerseyNum"} {
		t.Run(header, func(t *testing.T) {
			csv := "Team,Position,First Name,Last Name," + header + ",Overall\n" +
				"Bills,QB,Josh,Allen,17,91\n"
			rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
			require.NoError(t, err)
			require.Len(t, rows, 1)
			assert.True(t, hasTeam)
			assert.Equal(t, "17", rows[0].Number)
		})
	}
}

func TestParseCSV_RatingAliasVariants(t *testing.T) {
	for _, header := range []string{"Overall", "Overall Rating", "OverallRating", "OVR", "OVERALL RATING"} {
		t.Run(header, func(t *testing.T) {
			csv := "Team,Position,First Name,Last Name," + header + "\n" +
				"Bills,QB,Josh,Allen,91\n"
			rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
			require.NoError(t, err)
			require.Len(t, rows, 1)
			assert.True(t, hasTeam)
			assert.Equal(t, "91", rows[0].Rating)
		})
	}
}

func TestParseCSV_EmptyJerseyMeansKeepExisting(t *testing.T) {
	csv := "Team,Position,First Name,Last Name,Overall\n" +
		"Bills,QB,Josh,Allen,91\n"
	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, hasTeam)
	assert.Empty(t, rows[0].Number, "Jersey列自体が無いので空のまま")
}

func TestSplitDisplayName(t *testing.T) {
	t.Run("plain name", func(t *testing.T) {
		first, last := splitDisplayName("Josh Allen")
		assert.Equal(t, "Josh", first)
		assert.Equal(t, "Allen", last)
	})
	t.Run("suffix is stripped so it matches players.lastname from the roster scrape", func(t *testing.T) {
		first, last := splitDisplayName("Odell Beckham Jr.")
		assert.Equal(t, "Odell", first)
		assert.Equal(t, "Beckham", last)
	})
	t.Run("suffix without period", func(t *testing.T) {
		first, last := splitDisplayName("Antoine Winfield Jr")
		assert.Equal(t, "Antoine", first)
		assert.Equal(t, "Winfield", last)
	})
	t.Run("III suffix", func(t *testing.T) {
		first, last := splitDisplayName("Jessie Bates III")
		assert.Equal(t, "Jessie", first)
		assert.Equal(t, "Bates", last)
	})
}

func TestParseCSV_SeparateLastNameColumnWithSuffix(t *testing.T) {
	// 2024年分の実データと同じ: firstName/lastName が別列で、lastName に世代呼称が含まれる。
	csv := "firstName,lastName,Team,Position,Overall\n" +
		"Jessie,Bates III,Bengals,FS,88\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.True(t, hasTeam)
	require.Len(t, rows, 1)
	assert.Equal(t, "Jessie", rows[0].FirstName)
	assert.Equal(t, "Bates", rows[0].LastName, "players.lastname と揃うよう世代呼称を取り除く")
}

func TestTeamHintFromFilename(t *testing.T) {
	t.Run("known abbreviation", func(t *testing.T) {
		team, ok := TeamHintFromFilename("/data/2014/atl_madden_nfl_15.csv")
		require.True(t, ok)
		assert.Equal(t, "atl", team.Abbr)
	})
	t.Run("unknown prefix", func(t *testing.T) {
		_, ok := TeamHintFromFilename("/data/2024/Madden 25 Ratings.xlsx - Sheet 1.csv")
		assert.False(t, ok)
	})
}

func TestParseCSV_2024Style_NestedStatsColumnsDoNotShadowOverall(t *testing.T) {
	// 2024年分の実データと同じ問題: "Overall" 列の後に "stats/overall/value" のような入れ子列が続く。
	// 後者に上書きされて rating が取れなくなるバグの再発防止テスト。
	csv := "firstName,lastName,Team,Position,Overall,jerseyNum,stats/overall/value,stats/overall/diff\n" +
		"Josh,Allen,Buffalo Bills,QB,92,17,92,0\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.True(t, hasTeam)
	require.Len(t, rows, 1)

	r := rows[0]
	assert.Equal(t, "92", r.Rating, "先に出現する Overall 列を使う。入れ子の stats/overall/* に上書きされない")
	assert.Equal(t, "17", r.Number)
	assert.Equal(t, map[string]string{"stats/overall/value": "92", "stats/overall/diff": "0"}, r.Extra)
}

func TestParseCSV_InitialsFirstNameKeepsPeriodsAsGiven(t *testing.T) {
	// PFR（players テーブルの元データ）はイニシャル名のピリオド有無が選手ごとに不統一
	// （"A.J. Brown" はピリオド有り、"AJ Terrell" は無し）なので、パーサー側では正規化しない。
	// 揺れの吸収は importer.findRoster が検索時に行う。
	csv := "Team,Position,First Name,Last Name,Overall\n" +
		"Falcons,CB,A.J.,Terrell,85\n"

	rows, hasTeam, err := ParseCSV(strings.NewReader(csv))
	require.NoError(t, err)
	require.True(t, hasTeam)
	require.Len(t, rows, 1)
	assert.Equal(t, "A.J.", rows[0].FirstName)
}
