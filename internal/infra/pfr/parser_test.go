package pfr

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func loadFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/buf_2024_roster.html")
	require.NoError(t, err)
	return string(b)
}

func findByName(rows []domain.RawRosterRow, last, first string) (domain.RawRosterRow, bool) {
	for _, r := range rows {
		if r.LastName == last && r.FirstName == first {
			return r, true
		}
	}
	return domain.RawRosterRow{}, false
}

func TestParseRoster(t *testing.T) {
	rows, err := ParseRoster(loadFixture(t))
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	t.Run("drafted player", func(t *testing.T) {
		r, ok := findByName(rows, "Allen", "Josh")
		require.True(t, ok)
		assert.Equal(t, domain.RawRosterRow{
			Number:     "17",
			Position:   "QB",
			FirstName:  "Josh",
			LastName:   "Allen",
			Birthday:   "19960521",
			WeightLb:   "237",
			Height:     "6-5",
			College:    "Reedley", // 複数校ある場合は PHP 実装と同じく最初の <a> を採用
			Experience: "6",
			DraftTeam:  "Buffalo Bills",
			DraftRound: "1st",
			DraftRank:  "7th pick",
			DraftYear:  "2018",
		}, r)
	})

	t.Run("rookie without draft info", func(t *testing.T) {
		r, ok := findByName(rows, "Andreessen", "Joe")
		require.True(t, ok)
		assert.Equal(t, "0", r.Experience, "Rook は 0 に変換する")
		assert.Empty(t, r.DraftTeam)
		assert.Empty(t, r.DraftRound)
		assert.Empty(t, r.DraftRank)
		assert.Empty(t, r.DraftYear)
	})

	t.Run("every row has name, position and number", func(t *testing.T) {
		for _, r := range rows {
			assert.NotEmpty(t, r.LastName, "row %+v", r)
			assert.NotEmpty(t, r.Position)
		}
	})
}

func TestParseRoster_TableMissing(t *testing.T) {
	_, err := ParseRoster(`<html><body><table id="roster"></table></body></html>`)
	require.Error(t, err, ".is_setup が付いていない表は初期化前とみなして対象外")
}

func TestParseRoster_Fallbacks(t *testing.T) {
	html := `<div class="is_setup"><table id="roster"><tbody>
	<tr class="thead"><th data-stat="uniform_number">No.</th></tr>
	<tr>
	  <th data-stat="uniform_number"></th>
	  <td data-stat="player">John Doe Jr.</td>
	  <td data-stat="pos"></td>
	  <td data-stat="birth_date_mod"></td>
	  <td data-stat="college_id"></td>
	  <td data-stat="experience">Rook</td>
	  <td data-stat="draft_info">Team Only</td>
	</tr></tbody></table></div>`

	rows, err := ParseRoster(html)
	require.NoError(t, err)
	require.Len(t, rows, 1, "thead 行は除外する")

	r := rows[0]
	assert.Equal(t, "John Doe", r.FirstName, "csk 無しは表示名の最後の空白で分割")
	assert.Equal(t, "Jr.", r.LastName)
	assert.Equal(t, "No Data", r.Position)
	assert.Empty(t, r.Birthday)
	assert.Empty(t, r.College)
	assert.Equal(t, "0", r.Experience)
	assert.Equal(t, "Team Only", r.DraftTeam)
	assert.Empty(t, r.DraftRound)
	assert.Empty(t, r.DraftYear)
}
