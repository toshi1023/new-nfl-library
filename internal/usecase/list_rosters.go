package usecase

import (
	"context"
	"fmt"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// MaxTeamID は teams テーブルの id の最大値（32チーム固定）。
const MaxTeamID = 32

// ListRosters は season・team_id でロスターを検索するユースケース。
type ListRosters struct {
	repo domain.RosterRepository
}

// NewListRosters は ListRosters を生成する。
func NewListRosters(repo domain.RosterRepository) *ListRosters {
	return &ListRosters{repo: repo}
}

// Run は season・team_id でロスターを検索する。該当が無ければ空スライスを返す（エラーにしない）。
// season・team_id の範囲チェックは handler のクエリバインド（binding タグ）でも行っているが、
// usecase を直接呼ぶ別の経路（将来の拝備え）のためにここでも検証する。
func (u *ListRosters) Run(ctx context.Context, season, teamID int) ([]domain.RosterDetail, error) {
	if season < MinSeason {
		return nil, fmt.Errorf("season must be %d or later, got %d: %w", MinSeason, season, domain.ErrValidation)
	}
	if teamID < 1 || teamID > MaxTeamID {
		return nil, fmt.Errorf("team_id must be between 1 and %d, got %d: %w", MaxTeamID, teamID, domain.ErrValidation)
	}
	return u.repo.ListBySeasonAndTeam(ctx, season, teamID)
}
