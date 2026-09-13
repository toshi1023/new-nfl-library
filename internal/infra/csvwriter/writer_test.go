package csvwriter

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func f(v float64) *float64 { return &v }

func sample() []domain.RosterEntry {
	return []domain.RosterEntry{
		{
			Season: 2024, TeamID: 1, Team: "buf", Number: "17", Position: "QB",
			FirstName: "Josh", LastName: "Allen", Birthday: "19960521",
			WeightKg: f(106.7), HeightCm: f(195.6), College: "Wyoming", Experience: "6",
			DraftedTeam: "Buffalo Bills", DraftedRound: "1st", DraftedRank: "7th pick", DraftedYear: "2018",
		},
		{
			Season: 2024, TeamID: 1, Team: "buf", Position: "No Data",
			FirstName: "A, B", LastName: `C "D"`,
		},
	}
}

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, sample()))

	want := "season,team_id,team,number,position,firstname,lastname,birthday,weight_kg,height_cm,college,experience,drafted_team,drafted_round,drafted_rank,drafted_year\n" +
		"2024,1,buf,17,QB,Josh,Allen,19960521,106.7,195.6,Wyoming,6,Buffalo Bills,1st,7th pick,2018\n" +
		"2024,1,buf,,No Data,\"A, B\",\"C \"\"D\"\"\",,,,,,,,,\n"
	assert.Equal(t, want, buf.String())
}

func TestWriter_Write(t *testing.T) {
	// 存在しない中間ディレクトリも作成される
	path := filepath.Join(t.TempDir(), "outputs", "csv", "out.csv")
	w := NewWriter(path)
	require.NoError(t, w.Write(t.Context(), sample()))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, 3, bytes.Count(b, []byte("\n")), "ヘッダ + 2 行")
}
