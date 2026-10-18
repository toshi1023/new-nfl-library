package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func TestListTeams_Run(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo := &stubSearchRepository{teams: []domain.TeamDetail{{ID: 1, City: "Buffalo", Name: "Bills"}}}
		u := NewListTeams(repo)

		got, err := u.Run(t.Context())
		require.NoError(t, err)
		assert.Equal(t, repo.teams, got)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &stubSearchRepository{teamsErr: errors.New("db down")}
		u := NewListTeams(repo)

		_, err := u.Run(t.Context())
		require.Error(t, err)
	})
}
