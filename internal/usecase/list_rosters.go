package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// MaxTeamID は teams テーブルの id の最大値（32チーム固定）。
const MaxTeamID = 32

// DefaultSortBy・DefaultSortOrder は sort_by・sort_order 省略時の既定値。
const (
	DefaultSortBy    = "rating"
	DefaultSortOrder = "desc"
)

// validSortBy・validSortOrder は sort_by・sort_order に指定できる値。
var (
	validSortBy    = map[string]bool{"rating": true, "number": true, "experience": true, "name": true}
	validSortOrder = map[string]bool{"asc": true, "desc": true}
)

// ListRosters は season・team_id でロスターを検索するユースケース。
type ListRosters struct {
	repo domain.RosterRepository
}

// NewListRosters は ListRosters を生成する。
func NewListRosters(repo domain.RosterRepository) *ListRosters {
	return &ListRosters{repo: repo}
}

// Run は season・team_id でロスターを検索する。names が空でなければ選手名（firstname・lastname の
// いずれか）の部分一致でさらに絞り込む。複数指定時は OR 条件（いずれか1語でも一致すれば対象）。
// sortBy・sortOrder が空文字なら既定値（rating・desc）を使う。該当が無ければ空スライスを返す
// （エラーにしない）。
// season・team_id・sortBy・sortOrder の検証は handler のクエリバインド（binding タグ）でも行っているが、
// usecase を直接呼ぶ別の経路（将来の拝備え）のためにここでも検証する。
func (u *ListRosters) Run(ctx context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]domain.RosterDetail, error) {
	if season < MinSeason {
		return nil, fmt.Errorf("season must be %d or later, got %d: %w", MinSeason, season, domain.ErrValidation)
	}
	if teamID < 1 || teamID > MaxTeamID {
		return nil, fmt.Errorf("team_id must be between 1 and %d, got %d: %w", MaxTeamID, teamID, domain.ErrValidation)
	}
	if sortBy == "" {
		sortBy = DefaultSortBy
	}
	if sortOrder == "" {
		sortOrder = DefaultSortOrder
	}
	if !validSortBy[sortBy] {
		return nil, fmt.Errorf("sort_by must be one of rating/number/experience/name, got %q: %w", sortBy, domain.ErrValidation)
	}
	if !validSortOrder[sortOrder] {
		return nil, fmt.Errorf("sort_order must be one of asc/desc, got %q: %w", sortOrder, domain.ErrValidation)
	}
	return u.repo.ListBySeasonAndTeam(ctx, season, teamID, normalizeNames(names), sortBy, sortOrder)
}

// normalizeNames は各検索語の前後の空白を除き、空になった語を取り除く。
func normalizeNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		out = append(out, n)
	}
	return out
}
