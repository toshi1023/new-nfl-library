package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

type stubRosterRepository struct {
	rosters              []domain.RosterDetail
	err                  error
	gotSeason, gotTeamID int
}

func (s *stubRosterRepository) ListBySeasonAndTeam(_ context.Context, season, teamID int) ([]domain.RosterDetail, error) {
	s.gotSeason, s.gotTeamID = season, teamID
	return s.rosters, s.err
}

func TestListRosters_Run(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo := &stubRosterRepository{rosters: []domain.RosterDetail{{ID: 1}}}
		u := NewListRosters(repo)

		got, err := u.Run(t.Context(), 2024, 1)
		require.NoError(t, err)
		assert.Equal(t, repo.rosters, got)
		assert.Equal(t, 2024, repo.gotSeason)
		assert.Equal(t, 1, repo.gotTeamID)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &stubRosterRepository{err: errors.New("db down")}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1)
		require.Error(t, err)
	})

	t.Run("season before MinSeason is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2011, 1)
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("team_id out of range is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 0)
		require.ErrorIs(t, err, domain.ErrValidation)

		_, err = u.Run(t.Context(), 2024, 33)
		require.ErrorIs(t, err, domain.ErrValidation)
	})
}
