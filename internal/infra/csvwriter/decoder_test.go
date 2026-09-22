package csvwriter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeRosterEntries_RoundTrip(t *testing.T) {
	entries := sample()

	var buf strings.Builder
	require.NoError(t, Encode(&buf, entries))

	got, err := DecodeRosterEntries(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Equal(t, entries, got, "Encode したものを Decode すると元に戻る")
}

func TestDecodeRosterEntries_EmptyOptionalFields(t *testing.T) {
	csv := "season,team_id,team,number,position,firstname,lastname,birthday,weight_kg,height_cm,college,experience,drafted_team,drafted_round,drafted_rank,drafted_year\n" +
		"2024,1,buf,,No Data,A,B,,,,,,,,,\n"

	got, err := DecodeRosterEntries(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Nil(t, got[0].WeightKg)
	assert.Nil(t, got[0].HeightCm)
	assert.Empty(t, got[0].Number)
}

func TestDecodeRosterEntries_HeaderMismatch(t *testing.T) {
	_, err := DecodeRosterEntries(strings.NewReader("a,b,c\n1,2,3\n"))
	require.Error(t, err)
}

func TestDecodeRosterEntries_InvalidNumber(t *testing.T) {
	csv := "season,team_id,team,number,position,firstname,lastname,birthday,weight_kg,height_cm,college,experience,drafted_team,drafted_round,drafted_rank,drafted_year\n" +
		"abc,1,buf,,No Data,A,B,,,,,,,,,\n"
	_, err := DecodeRosterEntries(strings.NewReader(csv))
	require.Error(t, err)
}
