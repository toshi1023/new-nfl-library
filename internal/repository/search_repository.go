package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// SearchRepository は domain.SearchRepository の実装。
type SearchRepository struct {
	pool *pgxpool.Pool
}

// NewSearchRepository は SearchRepository を生成する。
func NewSearchRepository(pool *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{pool: pool}
}

// ListSeasons はロスターが存在するシーズンの一覧を降順で返す。
func (r *SearchRepository) ListSeasons(ctx context.Context) ([]int32, error) {
	seasons, err := sqlc.New(r.pool).ListRosterSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("list seasons: %w", err)
	}
	return seasons, nil
}

// ListTeams は全チームを id 昇順で返す。
func (r *SearchRepository) ListTeams(ctx context.Context) ([]domain.TeamDetail, error) {
	rows, err := sqlc.New(r.pool).ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	teams := make([]domain.TeamDetail, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, domain.TeamDetail{
			ID:            row.ID,
			City:          row.City,
			Name:          row.Name,
			Conference:    row.Conference,
			Area:          row.Area,
			ImageFile:     row.ImageFile,
			BackImageFile: row.BackImageFile,
		})
	}
	return teams, nil
}
