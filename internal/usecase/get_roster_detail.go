package usecase

import (
	"context"
	"fmt"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// GetRosterDetail はロスター1件の詳細（所属チーム・スターター情報・Madden 評価値・在籍履歴）を取得するユースケース。
type GetRosterDetail struct {
	repo domain.RosterRepository
}

// NewGetRosterDetail は GetRosterDetail を生成する。
func NewGetRosterDetail(repo domain.RosterRepository) *GetRosterDetail {
	return &GetRosterDetail{repo: repo}
}

// Run は rosters.id に一致するロスターの詳細を返す。在籍履歴（Career）には同じ選手の全シーズンの
// ロスターを含める。該当が無ければ domain.ErrNotFound をラップして返す。
// id の検証は handler のパスパラメータのバインド（binding タグ）でも行っているが、
// usecase を直接呼ぶ別の経路のためにここでも検証する。
// 2 つのクエリはいずれも読み取りのみで厳密な整合性を要求しないため、トランザクションは張らない。
func (u *GetRosterDetail) Run(ctx context.Context, id int) (domain.RosterProfile, error) {
	if id < 1 {
		return domain.RosterProfile{}, fmt.Errorf("id must be 1 or greater, got %d: %w", id, domain.ErrValidation)
	}

	profile, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.RosterProfile{}, fmt.Errorf("get roster: %w", err)
	}

	career, err := u.repo.ListCareerByPlayerID(ctx, int(profile.PlayerID))
	if err != nil {
		return domain.RosterProfile{}, fmt.Errorf("list roster career: %w", err)
	}
	profile.Career = career
	return profile, nil
}
