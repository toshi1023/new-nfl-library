package csvwriter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeStarterEntries_RoundTrip(t *testing.T) {
	entries := startersSample()

	var buf strings.Builder
	require.NoError(t, EncodeStarters(&buf, entries))

	got, err := DecodeStarterEntries(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Equal(t, entries, got, "EncodeStarters したものを Decode すると元に戻る")
}

func TestDecodeStarterEntries_HeaderMismatch(t *testing.T) {
	_, err := DecodeStarterEntries(strings.NewReader("a,b,c\n1,2,3\n"))
	require.Error(t, err)
}

func TestDecodeStarterEntries_InvalidOdflg(t *testing.T) {
	csv := "season,team_id,team,position,firstname,lastname,odflg\n" +
		"2024,1,buf,QB,Josh,Allen,maybe\n"
	_, err := DecodeStarterEntries(strings.NewReader(csv))
	require.Error(t, err)
}
