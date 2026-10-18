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
	rosters                 []domain.RosterDetail
	err                     error
	gotSeason, gotTeamID    int
	gotNames                []string
	gotSortBy, gotSortOrder string

	profile     domain.RosterProfile
	getErr      error
	gotID       int
	career      []domain.RosterCareer
	careerErr   error
	gotPlayerID int
}

func (s *stubRosterRepository) GetByID(_ context.Context, id int) (domain.RosterProfile, error) {
	s.gotID = id
	return s.profile, s.getErr
}

func (s *stubRosterRepository) ListCareerByPlayerID(_ context.Context, playerID int) ([]domain.RosterCareer, error) {
	s.gotPlayerID = playerID
	return s.career, s.careerErr
}

func (s *stubRosterRepository) ListBySeasonAndTeam(_ context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]domain.RosterDetail, error) {
	s.gotSeason, s.gotTeamID, s.gotNames = season, teamID, names
	s.gotSortBy, s.gotSortOrder = sortBy, sortOrder
	return s.rosters, s.err
}

func TestListRosters_Run(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo := &stubRosterRepository{rosters: []domain.RosterDetail{{ID: 1}}}
		u := NewListRosters(repo)

		got, err := u.Run(t.Context(), 2024, 1, nil, "", "")
		require.NoError(t, err)
		assert.Equal(t, repo.rosters, got)
		assert.Equal(t, 2024, repo.gotSeason)
		assert.Equal(t, 1, repo.gotTeamID)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &stubRosterRepository{err: errors.New("db down")}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "", "")
		require.Error(t, err)
	})

	t.Run("season before MinSeason is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2011, 1, nil, "", "")
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("team_id out of range is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 0, nil, "", "")
		require.ErrorIs(t, err, domain.ErrValidation)

		_, err = u.Run(t.Context(), 2024, 33, nil, "", "")
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("names are trimmed before reaching repository", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, []string{"  Josh  ", "  Allen  "}, "", "")
		require.NoError(t, err)
		assert.Equal(t, []string{"Josh", "Allen"}, repo.gotNames)
	})

	t.Run("blank names are dropped", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, []string{"   ", "", "Allen"}, "", "")
		require.NoError(t, err)
		assert.Equal(t, []string{"Allen"}, repo.gotNames)
	})

	t.Run("nil names stays empty (no filter)", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "", "")
		require.NoError(t, err)
		assert.Empty(t, repo.gotNames)
	})

	t.Run("sort_by/sort_order default to rating/desc when omitted", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "", "")
		require.NoError(t, err)
		assert.Equal(t, "rating", repo.gotSortBy)
		assert.Equal(t, "desc", repo.gotSortOrder)
	})

	t.Run("sort_by/sort_order pass through when given", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "name", "asc")
		require.NoError(t, err)
		assert.Equal(t, "name", repo.gotSortBy)
		assert.Equal(t, "asc", repo.gotSortOrder)
	})

	t.Run("invalid sort_by is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "bogus", "")
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("invalid sort_order is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewListRosters(repo)

		_, err := u.Run(t.Context(), 2024, 1, nil, "", "bogus")
		require.ErrorIs(t, err, domain.ErrValidation)
	})
}
