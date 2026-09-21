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

type stubStarterFetcher struct {
	rows  map[string][]domain.RawStarterRow
	fail  map[string]error
	calls []string
}

func (s *stubStarterFetcher) FetchStarters(_ context.Context, _ int, team domain.Team) ([]domain.RawStarterRow, error) {
	s.calls = append(s.calls, team.Abbr)
	if err, ok := s.fail[team.Abbr]; ok {
		return nil, err
	}
	return s.rows[team.Abbr], nil
}

type stubStarterWriter struct {
	got []domain.StarterEntry
	err error
}

func (s *stubStarterWriter) Write(_ context.Context, entries []domain.StarterEntry) error {
	s.got = entries
	return s.err
}

func TestScrapeStarters_Run(t *testing.T) {
	buf := domain.Team{ID: 1, Abbr: "buf"}
	nwe := domain.Team{ID: 2, Abbr: "nwe"}
	qb := domain.RawStarterRow{Position: "QB", FirstName: "Josh", LastName: "Allen"}
	de := domain.RawStarterRow{Position: "LDE", FirstName: "A.J.", LastName: "Epenesa"}

	t.Run("collects all teams then writes once, odflg from position category", func(t *testing.T) {
		fetcher := &stubStarterFetcher{rows: map[string][]domain.RawStarterRow{"buf": {qb, de}, "nwe": {qb}}}
		writer := &stubStarterWriter{}
		uc := NewScrapeStarters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf, nwe})
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, []string{"buf", "nwe"}, fetcher.calls)
		require.Len(t, writer.got, 3)

		assert.Equal(t, domain.StarterEntry{Season: 2024, TeamID: 1, Team: "buf", Position: "QB", FirstName: "Josh", LastName: "Allen", Odflg: true}, writer.got[0])
		assert.Equal(t, domain.StarterEntry{Season: 2024, TeamID: 1, Team: "buf", Position: "LDE", FirstName: "A.J.", LastName: "Epenesa", Odflg: false}, writer.got[1])
		assert.Equal(t, 2, writer.got[2].TeamID)
	})

	t.Run("unknown position falls back to odflg=false", func(t *testing.T) {
		fetcher := &stubStarterFetcher{rows: map[string][]domain.RawStarterRow{
			"buf": {{Position: "???", FirstName: "No", LastName: "Body"}},
		}}
		writer := &stubStarterWriter{}
		uc := NewScrapeStarters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf})
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.False(t, writer.got[0].Odflg)
	})

	t.Run("writes partial result when a team fails", func(t *testing.T) {
		boom := errors.New("boom")
		fetcher := &stubStarterFetcher{
			rows: map[string][]domain.RawStarterRow{"buf": {qb}},
			fail: map[string]error{"nwe": boom},
		}
		writer := &stubStarterWriter{}
		uc := NewScrapeStarters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf, nwe, {ID: 3, Abbr: "mia"}})
		require.ErrorIs(t, err, boom)
		assert.Equal(t, 1, n)
		assert.Equal(t, []string{"buf", "nwe"}, fetcher.calls, "失敗した時点で後続チームは取得しない")
		assert.Len(t, writer.got, 1, "取得済み分は書き出す")
	})

	t.Run("nothing written when first team fails", func(t *testing.T) {
		fetcher := &stubStarterFetcher{fail: map[string]error{"buf": errors.New("boom")}}
		writer := &stubStarterWriter{}
		uc := NewScrapeStarters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf})
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.Nil(t, writer.got)
	})

	t.Run("writer error is joined", func(t *testing.T) {
		werr := errors.New("disk full")
		fetcher := &stubStarterFetcher{rows: map[string][]domain.RawStarterRow{"buf": {qb}}}
		writer := &stubStarterWriter{err: werr}
		uc := NewScrapeStarters(fetcher, writer, 0)

		n, err := uc.Run(t.Context(), 2024, []domain.Team{buf})
		require.ErrorIs(t, err, werr)
		assert.Equal(t, 0, n)
	})

	t.Run("season before 2012 is rejected", func(t *testing.T) {
		uc := NewScrapeStarters(&stubStarterFetcher{}, &stubStarterWriter{}, 0)
		_, err := uc.Run(t.Context(), 2011, []domain.Team{buf})
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("cancel during delay stops and writes partial", func(t *testing.T) {
		fetcher := &stubStarterFetcher{rows: map[string][]domain.RawStarterRow{"buf": {qb}, "nwe": {qb}}}
		writer := &stubStarterWriter{}
		uc := NewScrapeStarters(fetcher, writer, time.Minute)

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
