package usecase

import (
	"context"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// ListSeasons は検索UI用に、ロスターが存在するシーズン一覧を返すユースケース。
type ListSeasons struct {
	repo domain.SearchRepository
}

// NewListSeasons は ListSeasons を生成する。
func NewListSeasons(repo domain.SearchRepository) *ListSeasons {
	return &ListSeasons{repo: repo}
}

// Run はシーズン一覧を降順で返す。
func (u *ListSeasons) Run(ctx context.Context) ([]int32, error) {
	return u.repo.ListSeasons(ctx)
}
