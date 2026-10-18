package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func TestGetRosterDetail_Run(t *testing.T) {
	t.Run("returns profile with career of the same player", func(t *testing.T) {
		repo := &stubRosterRepository{
			profile: domain.RosterProfile{RosterDetail: domain.RosterDetail{ID: 5, PlayerID: 10}},
			career:  []domain.RosterCareer{{RosterID: 5, Season: 2024}, {RosterID: 3, Season: 2023}},
		}
		u := NewGetRosterDetail(repo)

		got, err := u.Run(t.Context(), 5)
		require.NoError(t, err)
		assert.Equal(t, 5, repo.gotID)
		assert.Equal(t, 10, repo.gotPlayerID, "在籍履歴は取得した roster の player_id で引く")
		assert.EqualValues(t, 5, got.ID)
		assert.Equal(t, repo.career, got.Career)
	})

	t.Run("id less than 1 is rejected", func(t *testing.T) {
		repo := &stubRosterRepository{}
		u := NewGetRosterDetail(repo)

		_, err := u.Run(t.Context(), 0)
		require.ErrorIs(t, err, domain.ErrValidation)
		assert.Zero(t, repo.gotID, "不正な id では repository を呼ばない")
	})

	t.Run("propagates not found", func(t *testing.T) {
		repo := &stubRosterRepository{getErr: domain.ErrNotFound}
		u := NewGetRosterDetail(repo)

		_, err := u.Run(t.Context(), 999)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Zero(t, repo.gotPlayerID, "roster が無ければ在籍履歴は取得しない")
	})

	t.Run("propagates career error", func(t *testing.T) {
		repo := &stubRosterRepository{
			profile:   domain.RosterProfile{RosterDetail: domain.RosterDetail{ID: 5, PlayerID: 10}},
			careerErr: errors.New("db down"),
		}
		u := NewGetRosterDetail(repo)

		_, err := u.Run(t.Context(), 5)
		require.Error(t, err)
	})
}
