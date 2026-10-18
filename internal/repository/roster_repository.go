// Package repository は domain のリポジトリインターフェースの実装を提供する。
// sqlc 生成コード（internal/repository/sqlc）をラップし、domain のエンティティに変換する。
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// RosterRepository は domain.RosterRepository の実装。
type RosterRepository struct {
	pool *pgxpool.Pool
}

// NewRosterRepository は RosterRepository を生成する。
func NewRosterRepository(pool *pgxpool.Pool) *RosterRepository {
	return &RosterRepository{pool: pool}
}

// ListBySeasonAndTeam は season・team_id に一致する roster を player・position と合わせて返す。
// names が空でなければ、いずれか1語でも firstname・lastname に部分一致する roster に絞り込む（OR条件）。
// sortBy・sortOrder で並び順を指定する（domain.RosterRepository のコメント参照）。
func (r *RosterRepository) ListBySeasonAndTeam(ctx context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]domain.RosterDetail, error) {
	escaped := make([]string, 0, len(names))
	for _, n := range names {
		escaped = append(escaped, escapeLikePattern(n))
	}

	q := sqlc.New(r.pool)
	rows, err := q.ListRostersBySeasonTeam(ctx, sqlc.ListRostersBySeasonTeamParams{
		Season:   int32(season),
		TeamID:   int32(teamID),
		Names:    escaped,
		SortBy:   sortBy,
		SortDesc: sortOrder == "desc",
	})
	if err != nil {
		return nil, fmt.Errorf("list rosters: %w", err)
	}

	details := make([]domain.RosterDetail, 0, len(rows))
	for _, row := range rows {
		d, err := toRosterDetail(row)
		if err != nil {
			return nil, fmt.Errorf("roster id %d: %w", row.ID, err)
		}
		details = append(details, d)
	}
	return details, nil
}

func toRosterDetail(row *sqlc.ListRostersBySeasonTeamRow) (domain.RosterDetail, error) {
	maddenAttributes, err := decodeMaddenAttributes(row.MaddenAttributes)
	if err != nil {
		return domain.RosterDetail{}, fmt.Errorf("decode madden_attributes: %w", err)
	}

	height, err := numericToFloat(row.Height)
	if err != nil {
		return domain.RosterDetail{}, fmt.Errorf("height: %w", err)
	}
	weight, err := numericToFloat(row.Weight)
	if err != nil {
		return domain.RosterDetail{}, fmt.Errorf("weight: %w", err)
	}

	return domain.RosterDetail{
		ID:         row.ID,
		Season:     row.Season,
		TeamID:     row.TeamID,
		PlayerID:   row.PlayerID,
		PositionID: row.PositionID,
		Number:     row.Number,
		Rating:     row.Rating,
		Experience: row.Experience,
		Player: domain.Player{
			ID:           row.PlayerTableID,
			FirstName:    row.Firstname,
			LastName:     row.Lastname,
			Birthday:     row.Birthday,
			HeightCm:     height,
			WeightKg:     weight,
			College:      row.College,
			DraftedTeam:  row.DraftedTeam,
			DraftedRound: row.DraftedRound,
			DraftedRank:  row.DraftedRank,
			DraftedYear:  row.DraftedYear,
			ImageFile:    row.ImageFile,
		},
		Position: domain.PositionDetail{
			ID:               row.PositionTableID,
			AbstractCategory: row.AbstractCategory,
			Category:         row.Category,
			Name:             row.PositionName,
			Odflg:            row.Odflg,
		},
		MaddenAttributes: maddenAttributes,
	}, nil
}

// decodeMaddenAttributes は rosters.madden_attributes（JSONB の生バイト列）を map に変換する。
// 未投入（NULL）の場合は nil を返す。
func decodeMaddenAttributes(raw []byte) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// escapeLikePattern は ILIKE のワイルドカード文字（\, %, _）をエスケープする。
// 選手名のフリー入力をそのまま '%' || name || '%' に埋め込むため、入力に % や _ が含まれていても
// 意図しないワイルドカードとして展開されないようにする（クエリ側は ESCAPE '\' を指定している）。
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}
