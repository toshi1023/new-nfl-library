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

func startersSample() []domain.StarterEntry {
	return []domain.StarterEntry{
		{Season: 2024, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Josh", LastName: "Allen", Odflg: true},
		{Season: 2024, TeamID: 1, Team: "buf", Position: "LDE", FirstName: "A.J.", LastName: "Epenesa", Odflg: false},
	}
}

func TestEncodeStarters(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, EncodeStarters(&buf, startersSample()))

	want := "season,team_id,team,position,firstname,lastname,odflg\n" +
		"2024,1,buf,QB,Josh,Allen,true\n" +
		"2024,1,buf,LDE,A.J.,Epenesa,false\n"
	assert.Equal(t, want, buf.String())
}

func TestStartersWriter_Write(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outputs", "csv", "out.csv")
	w := NewStartersWriter(path)
	require.NoError(t, w.Write(t.Context(), startersSample()))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, 3, bytes.Count(b, []byte("\n")), "ヘッダ + 2 行")
}
