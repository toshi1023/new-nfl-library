// Package repository は domain のリポジトリインターフェースの実装を提供する。
// sqlc 生成コード（internal/repository/sqlc）をラップし、domain のエンティティに変換する。
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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

// GetByID は rosters.id に一致する roster を player・position・team・starter と合わせて返す。
// Career は含まない（ListCareerByPlayerID で別途取得する）。
func (r *RosterRepository) GetByID(ctx context.Context, id int) (domain.RosterProfile, error) {
	row, err := sqlc.New(r.pool).GetRosterDetailByID(ctx, int32(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RosterProfile{}, fmt.Errorf("roster id %d: %w", id, domain.ErrNotFound)
	}
	if err != nil {
		return domain.RosterProfile{}, fmt.Errorf("get roster id %d: %w", id, err)
	}

	maddenAttributes, err := decodeMaddenAttributes(row.MaddenAttributes)
	if err != nil {
		return domain.RosterProfile{}, fmt.Errorf("roster id %d: decode madden_attributes: %w", id, err)
	}
	player, err := toPlayer(row.Player)
	if err != nil {
		return domain.RosterProfile{}, fmt.Errorf("roster id %d: %w", id, err)
	}

	var starter *domain.RosterStarter
	if row.StarterID != nil {
		starter = &domain.RosterStarter{Odflg: row.StarterOdflg != nil && *row.StarterOdflg}
	}

	return domain.RosterProfile{
		RosterDetail: domain.RosterDetail{
			ID:         row.ID,
			Season:     row.Season,
			TeamID:     row.TeamID,
			PlayerID:   row.PlayerID,
			PositionID: row.PositionID,
			Number:     row.Number,
			Rating:     row.Rating,
			Experience: row.Experience,
			Player:     player,
			Position:   toPositionDetail(row.Position),
		},
		Team:             toTeamDetail(row.Team),
		Starter:          starter,
		MaddenAttributes: maddenAttributes,
	}, nil
}

// ListCareerByPlayerID は player_id に一致する全シーズンの roster を team と合わせて season の降順で返す。
func (r *RosterRepository) ListCareerByPlayerID(ctx context.Context, playerID int) ([]domain.RosterCareer, error) {
	rows, err := sqlc.New(r.pool).ListRosterCareerByPlayerID(ctx, int32(playerID))
	if err != nil {
		return nil, fmt.Errorf("list roster career of player id %d: %w", playerID, err)
	}

	careers := make([]domain.RosterCareer, 0, len(rows))
	for _, row := range rows {
		careers = append(careers, domain.RosterCareer{
			RosterID:     row.ID,
			Season:       row.Season,
			Team:         toTeamDetail(row.Team),
			PositionName: row.PositionName,
			Number:       row.Number,
			Rating:       row.Rating,
			Experience:   row.Experience,
			IsStarter:    row.IsStarter,
		})
	}
	return careers, nil
}

func toRosterDetail(row *sqlc.ListRostersBySeasonTeamRow) (domain.RosterDetail, error) {
	player, err := toPlayer(row.Player)
	if err != nil {
		return domain.RosterDetail{}, err
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
		Player:     player,
		Position:   toPositionDetail(row.Position),
	}, nil
}

func toPlayer(p sqlc.Player) (domain.Player, error) {
	height, err := numericToFloat(p.Height)
	if err != nil {
		return domain.Player{}, fmt.Errorf("height: %w", err)
	}
	weight, err := numericToFloat(p.Weight)
	if err != nil {
		return domain.Player{}, fmt.Errorf("weight: %w", err)
	}

	return domain.Player{
		ID:           p.ID,
		FirstName:    p.Firstname,
		LastName:     p.Lastname,
		Birthday:     p.Birthday,
		HeightCm:     height,
		WeightKg:     weight,
		College:      p.College,
		DraftedTeam:  p.DraftedTeam,
		DraftedRound: p.DraftedRound,
		DraftedRank:  p.DraftedRank,
		DraftedYear:  p.DraftedYear,
		ImageFile:    p.ImageFile,
	}, nil
}

func toPositionDetail(p sqlc.Position) domain.PositionDetail {
	return domain.PositionDetail{
		ID:               p.ID,
		AbstractCategory: p.AbstractCategory,
		Category:         p.Category,
		Name:             p.Name,
		Odflg:            p.Odflg,
	}
}

func toTeamDetail(t sqlc.Team) domain.TeamDetail {
	return domain.TeamDetail{
		ID:            t.ID,
		City:          t.City,
		Name:          t.Name,
		Conference:    t.Conference,
		Area:          t.Area,
		ImageFile:     t.ImageFile,
		BackImageFile: t.BackImageFile,
	}
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
