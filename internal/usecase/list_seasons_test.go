package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// stubSearchRepository は domain.SearchRepository のテスト用スタブ。
// list_seasons_test.go / list_teams_test.go の両方から使う。
type stubSearchRepository struct {
	seasons    []int32
	seasonsErr error
	teams      []domain.TeamDetail
	teamsErr   error
}

func (s *stubSearchRepository) ListSeasons(_ context.Context) ([]int32, error) {
	return s.seasons, s.seasonsErr
}

func (s *stubSearchRepository) ListTeams(_ context.Context) ([]domain.TeamDetail, error) {
	return s.teams, s.teamsErr
}

func TestListSeasons_Run(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo := &stubSearchRepository{seasons: []int32{2025, 2024, 2023}}
		u := NewListSeasons(repo)

		got, err := u.Run(t.Context())
		require.NoError(t, err)
		assert.Equal(t, repo.seasons, got)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &stubSearchRepository{seasonsErr: errors.New("db down")}
		u := NewListSeasons(repo)

		_, err := u.Run(t.Context())
		require.Error(t, err)
	})
}
