package pfr

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func loadStartersFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/buf_2024_starters.html")
	require.NoError(t, err)
	return string(b)
}

func TestParseStarters(t *testing.T) {
	rows, err := ParseStarters(loadStartersFixture(t))
	require.NoError(t, err)
	require.Len(t, rows, 22, "オフェンス11人 + ディフェンス11人。見出し行（Offensive/Defensive Starters）は除外される")

	assert.Equal(t, domain.RawStarterRow{Position: "QB", FirstName: "Josh", LastName: "Allen"}, rows[0])
	assert.Equal(t, domain.RawStarterRow{Position: "RB", FirstName: "James", LastName: "Cook"}, rows[1])
	assert.Equal(t, domain.RawStarterRow{Position: "LDE", FirstName: "A.J.", LastName: "Epenesa"}, rows[11], "ディフェンス側の先頭")
	assert.Equal(t, domain.RawStarterRow{Position: "DB", FirstName: "Taron", LastName: "Johnson"}, rows[21])

	for _, r := range rows {
		assert.NotEmpty(t, r.Position)
		assert.NotEmpty(t, r.LastName)
	}
}

func TestParseStarters_TableMissing(t *testing.T) {
	_, err := ParseStarters(`<html><body><table id="roster"></table></body></html>`)
	require.Error(t, err)
}

func TestParseStarters_Fallback(t *testing.T) {
	html := `<table id="starters"><tbody>
	<tr><th data-stat="pos"></th><td data-stat="player"><b>Offensive Starters</b></td></tr>
	<tr class="full_table"><th data-stat="pos">QB</th><td data-stat="player">John Doe Jr.</td></tr>
	</tbody></table>`

	rows, err := ParseStarters(html)
	require.NoError(t, err)
	require.Len(t, rows, 1, "見出し行（full_table クラス無し）は除外する")
	assert.Equal(t, "QB", rows[0].Position)
	assert.Equal(t, "John Doe", rows[0].FirstName, "csk 無しは表示名の最後の空白で分割")
	assert.Equal(t, "Jr.", rows[0].LastName)
}
