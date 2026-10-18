package usecase

import (
	"context"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// ListTeams は検索UI用に、全チーム一覧を返すユースケース。
type ListTeams struct {
	repo domain.SearchRepository
}

// NewListTeams は ListTeams を生成する。
func NewListTeams(repo domain.SearchRepository) *ListTeams {
	return &ListTeams{repo: repo}
}

// Run はチーム一覧を id 昇順で返す。
func (u *ListTeams) Run(ctx context.Context) ([]domain.TeamDetail, error) {
	return u.repo.ListTeams(ctx)
}
