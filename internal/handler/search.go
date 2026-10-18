package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// SeasonLister は検索UI用にロスターが存在するシーズン一覧を返す（internal/usecase.ListSeasons が実装）。
type SeasonLister interface {
	Run(ctx context.Context) ([]int32, error)
}

// TeamLister は検索UI用にチーム一覧を返す（internal/usecase.ListTeams が実装）。
type TeamLister interface {
	Run(ctx context.Context) ([]domain.TeamDetail, error)
}

// SearchHandler は GET /api/v1/search/seasons・GET /api/v1/search/teams を提供する。
type SearchHandler struct {
	seasons SeasonLister
	teams   TeamLister
}

// NewSearchHandler は SearchHandler を生成する。
func NewSearchHandler(seasons SeasonLister, teams TeamLister) *SearchHandler {
	return &SearchHandler{seasons: seasons, teams: teams}
}

// seasonResponse は Flutter 側の SeasonResponse（lib/types/api/search/season_response.dart）と一致させる。
type seasonResponse struct {
	Season int32 `json:"season"`
}

// teamResponse は Flutter 側の TeamResponse（lib/types/api/search/team_response.dart）と一致させる。
// conference・area は DB 上は数値（1/2、1〜4）だが、Dart 側は表示用の文字列を期待するためここで変換する。
type teamResponse struct {
	ID            int32   `json:"id"`
	City          string  `json:"city"`
	Name          string  `json:"name"`
	Conference    string  `json:"conference"`
	Area          string  `json:"area"`
	ImageFile     *string `json:"image_file"`
	BackImageFile *string `json:"back_image_file"`
}

// Seasons は GET /api/v1/search/seasons を処理する。
func (h *SearchHandler) Seasons(c *gin.Context) {
	seasons, err := h.seasons.Run(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]seasonResponse, 0, len(seasons))
	for _, s := range seasons {
		out = append(out, seasonResponse{Season: s})
	}
	respondSuccess(c, "シーズン一覧を取得しました", gin.H{"seasons": out})
}

// Teams は GET /api/v1/search/teams を処理する。
func (h *SearchHandler) Teams(c *gin.Context) {
	teams, err := h.teams.Run(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]teamResponse, 0, len(teams))
	for _, t := range teams {
		out = append(out, toTeamResponse(t))
	}
	respondSuccess(c, "チーム一覧を取得しました", gin.H{"teams": out})
}

func toTeamResponse(t domain.TeamDetail) teamResponse {
	return teamResponse{
		ID:            t.ID,
		City:          t.City,
		Name:          t.Name,
		Conference:    conferenceLabel(t.Conference),
		Area:          areaLabel(t.Area),
		ImageFile:     t.ImageFile,
		BackImageFile: t.BackImageFile,
	}
}

// conferenceLabel は teams.conference（1: AFC, 2: NFC）を表示用文字列に変換する。
func conferenceLabel(c int16) string {
	switch c {
	case 1:
		return "AFC"
	case 2:
		return "NFC"
	default:
		return strconv.Itoa(int(c))
	}
}

// areaLabel は teams.area（1: 東, 2: 北, 3: 南, 4: 西）を表示用文字列に変換する。
func areaLabel(a int16) string {
	switch a {
	case 1:
		return "東"
	case 2:
		return "北"
	case 3:
		return "南"
	case 4:
		return "西"
	default:
		return strconv.Itoa(int(a))
	}
}
