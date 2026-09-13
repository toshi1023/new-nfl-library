package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

type stubFetcher struct {
	rows  map[string][]domain.RawRosterRow
	fail  map[string]error
	calls []string
}

func (s *stubFetcher) FetchRoster(_ context.Context, _ int, team domain.Team) ([]domain.RawRosterRow, error) {
	s.calls = append(s.calls, team.Abbr)
	if err, ok := s.fail[team.Abbr]; ok {
		return nil, err
	}
	return s.rows[team.Abbr], nil
}

type stubWriter struct {
	got []domain.RosterEntry
	err error
}

func (s *stubWriter) Write(_ context.Context, entries []domain.RosterEntry) error {
	s.got = entries
	return s.err
}

func TestScrapeRosters_Run(t *testing.T) {
	buf := domain.Team{ID: 1, Abbr: "buf"}
	nwe := domain.Team{ID: 2, Abbr: "nwe"}
	row := domain.RawRosterRow{FirstName: "Josh", LastName: "Allen", WeightLb: "237", Height: "6-5", Position: "QB"}

	t.Run("collects all teams then writes once", func(t *testing.T) {
		fetcher := &stubFetcher{rows: map[string][]domain.RawRosterRow{"buf": {row, row}, "nwe": {row}}}
		writer := &stubWriter{}
		uc := NewScrapeRosters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf, nwe})
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, []string{"buf", "nwe"}, fetcher.calls)
		require.Len(t, writer.got, 3)

		e := writer.got[0]
		assert.Equal(t, 2024, e.Season)
		assert.Equal(t, 1, e.TeamID)
		assert.Equal(t, "buf", e.Team)
		assert.Equal(t, "Allen", e.LastName)
		require.NotNil(t, e.WeightKg)
		assert.InDelta(t, 106.7, *e.WeightKg, 0.001, "237lb * 0.45")
		require.NotNil(t, e.HeightCm)
		assert.InDelta(t, 195.6, *e.HeightCm, 0.001, "6*30.48 + 5*2.54 = 195.58")
		assert.Equal(t, 2, writer.got[2].TeamID)
	})

	t.Run("writes partial result when a team fails", func(t *testing.T) {
		boom := errors.New("boom")
		fetcher := &stubFetcher{
			rows: map[string][]domain.RawRosterRow{"buf": {row}},
			fail: map[string]error{"nwe": boom},
		}
		writer := &stubWriter{}
		uc := NewScrapeRosters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf, nwe, {ID: 3, Abbr: "mia"}})
		require.ErrorIs(t, err, boom)
		assert.Equal(t, 1, n)
		assert.Equal(t, []string{"buf", "nwe"}, fetcher.calls, "失敗した時点で後続チームは取得しない")
		assert.Len(t, writer.got, 1, "取得済み分は書き出す")
	})

	t.Run("nothing written when first team fails", func(t *testing.T) {
		fetcher := &stubFetcher{fail: map[string]error{"buf": errors.New("boom")}}
		writer := &stubWriter{}
		uc := NewScrapeRosters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf})
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.Nil(t, writer.got)
	})

	t.Run("writer error is joined", func(t *testing.T) {
		werr := errors.New("disk full")
		fetcher := &stubFetcher{rows: map[string][]domain.RawRosterRow{"buf": {row}}}
		writer := &stubWriter{err: werr}
		uc := NewScrapeRosters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf})
		require.ErrorIs(t, err, werr)
		assert.Equal(t, 0, n)
	})

	t.Run("season before 2012 is rejected", func(t *testing.T) {
		uc := NewScrapeRosters(&stubFetcher{}, &stubWriter{}, 0)
		_, err := uc.Run(t.Context(), 2011, []domain.Team{buf})
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("cancel during delay stops and writes partial", func(t *testing.T) {
		fetcher := &stubFetcher{rows: map[string][]domain.RawRosterRow{"buf": {row}, "nwe": {row}}}
		writer := &stubWriter{}
		uc := NewScrapeRosters(fetcher, writer, time.Minute)

		ctx, cancel := context.WithCancel(t.Context())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		n, err := uc.Run(ctx, 2024, []domain.Team{buf, nwe})
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 1, n)
		assert.Equal(t, []string{"buf"}, fetcher.calls)
	})
}

func TestConvert(t *testing.T) {
	assert.Nil(t, convertWeight(""))
	assert.Nil(t, convertWeight("abc"))
	assert.Nil(t, convertWeight("0"))
	assert.InDelta(t, 90.0, *convertWeight("200"), 0.001)

	assert.Nil(t, convertHeight(""))
	assert.Nil(t, convertHeight("6"))
	assert.Nil(t, convertHeight("a-b"))
	assert.InDelta(t, 182.9, *convertHeight("6-0"), 0.001)
	assert.InDelta(t, 188.0, *convertHeight("6-2"), 0.001, "6*30.48 + 2*2.54 = 187.96 → 四捨五入で 188.0")
}
