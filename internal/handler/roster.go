package handler

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/maddenimport"
)

// RosterLister は season・team_id・選手名でロスターを検索する（internal/usecase.ListRosters が実装）。
type RosterLister interface {
	Run(ctx context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]domain.RosterDetail, error)
}

// RosterDetailGetter はロスター1件の詳細を取得する（internal/usecase.GetRosterDetail が実装）。
type RosterDetailGetter interface {
	Run(ctx context.Context, id int) (domain.RosterProfile, error)
}

// RosterHandler は GET /api/v1/rosters・GET /api/v1/rosters/{id} を提供する。
type RosterHandler struct {
	lister RosterLister
	getter RosterDetailGetter
	// assetBaseURL は選手画像の配信元URL。空文字なら image_file があっても image_url は組み立てない
	// （画像配信先がまだ決まっていないため。環境変数 ASSET_BASE_URL で設定する）。
	assetBaseURL string
}

// NewRosterHandler は RosterHandler を生成する。
func NewRosterHandler(lister RosterLister, getter RosterDetailGetter, assetBaseURL string) *RosterHandler {
	return &RosterHandler{lister: lister, getter: getter, assetBaseURL: assetBaseURL}
}

// rostersQuery は GET /api/v1/rosters のクエリパラメータ。Name は任意（選手名のフリー入力検索）。
// ?name=Josh&name=Allen のような繰り返し指定と、?name=Josh,Allen のようなカンマ区切りの
// 1つの値のどちらでも複数語を受け付ける（splitNames で正規化する）。
// SortBy・SortOrder も任意。空文字なら usecase 側で既定値（rating・desc）を使う。
type rostersQuery struct {
	Season    int      `form:"season" binding:"required,min=2012"`
	TeamID    int      `form:"team_id" binding:"required,min=1,max=32"`
	Name      []string `form:"name"`
	SortBy    string   `form:"sort_by" binding:"omitempty,oneof=rating number experience name"`
	SortOrder string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// rosterURI は GET /api/v1/rosters/{id} のパスパラメータ。
type rosterURI struct {
	ID int `uri:"id" binding:"required,min=1"`
}

// rosterResponse は /api/v1/rosters のレスポンス 1 件分。
type rosterResponse struct {
	ID         int32                  `json:"id"`
	Season     int32                  `json:"season"`
	TeamID     int32                  `json:"team_id"`
	PlayerID   int32                  `json:"player_id"`
	PositionID int32                  `json:"position_id"`
	Number     *int32                 `json:"number"`
	Rating     *int32                 `json:"rating"`
	Experience *int32                 `json:"experience"`
	Player     rosterPlayerResponse   `json:"player"`
	Position   rosterPositionResponse `json:"position"`
}

// rosterDetailResponse は /api/v1/rosters/{id} のレスポンス。一覧の rosterResponse の項目
// （埋め込みでフラットに展開される）に所属チーム・スターター情報・Madden 評価値・在籍履歴を加えたもの。
type rosterDetailResponse struct {
	rosterResponse
	Team             teamResponse              `json:"team"`
	Starter          *rosterStarterResponse    `json:"starter"`
	MaddenAttributes []maddenAttributeResponse `json:"madden_attributes"`
	Career           []rosterCareerResponse    `json:"career"`
}

// rosterStarterResponse はスターター情報。odflg は rosterPositionResponse と同じく 0/1 で返す。
type rosterStarterResponse struct {
	Odflg int `json:"odflg"`
}

// rosterCareerResponse は在籍履歴1件。
type rosterCareerResponse struct {
	RosterID     int32        `json:"roster_id"`
	Season       int32        `json:"season"`
	Team         teamResponse `json:"team"`
	PositionName string       `json:"position_name"`
	Number       *int32       `json:"number"`
	Rating       *int32       `json:"rating"`
	Experience   *int32       `json:"experience"`
	IsStarter    bool         `json:"is_starter"`
}

// rosterPlayerResponse は Flutter 側の RosterPlayerResponse（lib/types/api/roster/roster_response.dart）
// と一致させる。birthday_year・birthday_date・image_url は Laravel の Player モデルのアクセサ相当を
// ここで算出する（domain.Player には持たせず、表示用の変換として handler の責務にする）。
type rosterPlayerResponse struct {
	ID           int32    `json:"id"`
	FirstName    string   `json:"firstname"`
	LastName     string   `json:"lastname"`
	Birthday     string   `json:"birthday"`
	Height       *float64 `json:"height"`
	Weight       *float64 `json:"weight"`
	College      *string  `json:"college"`
	DraftedTeam  *string  `json:"drafted_team"`
	DraftedRound *string  `json:"drafted_round"`
	DraftedRank  *string  `json:"drafted_rank"`
	DraftedYear  *int32   `json:"drafted_year"`
	ImageFile    *string  `json:"image_file"`
	BirthdayYear *int     `json:"birthday_year"`
	BirthdayDate *string  `json:"birthday_date"`
	ImageURL     *string  `json:"image_url"`
}

// rosterPositionResponse は Flutter 側の RosterPositionResponse と一致させる。
// odflg は Dart 側が int? 定義のため bool を 0/1 に変換する。
type rosterPositionResponse struct {
	ID               int32  `json:"id"`
	AbstractCategory *int32 `json:"abstract_category"`
	Category         *int32 `json:"category"`
	Name             string `json:"name"`
	Odflg            *int   `json:"odflg"`
}

// maddenAttributeResponse は madden_attributes の1項目。生キー・日本語名・値をセットで返す。
type maddenAttributeResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// Index は GET /api/v1/rosters を処理する。
func (h *RosterHandler) Index(c *gin.Context) {
	var q rostersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respondBindError(c, err)
		return
	}

	rosters, err := h.lister.Run(c.Request.Context(), q.Season, q.TeamID, splitNames(q.Name), q.SortBy, q.SortOrder)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, "ロスター情報を取得しました", gin.H{
		"rosters": h.toRosterResponses(rosters),
	})
}

// Show は GET /api/v1/rosters/{id} を処理する。
func (h *RosterHandler) Show(c *gin.Context) {
	var uri rosterURI
	if err := c.ShouldBindUri(&uri); err != nil {
		respondBindError(c, err)
		return
	}

	profile, err := h.getter.Run(c.Request.Context(), uri.ID)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, "ロスター詳細を取得しました", gin.H{
		"roster": h.toRosterDetailResponse(profile),
	})
}

// splitNames は "name" クエリパラメータをフラットな検索語のリストに変換する。
// ?name=Josh&name=Allen（繰り返し指定）と ?name=Josh,Allen（カンマ区切りの1つの値）の
// どちらも ["Josh", "Allen"] になる（両方を組み合わせてもよい）。前後の空白は除き、空の語は捨てる。
func splitNames(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		for _, part := range strings.Split(r, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			out = append(out, part)
		}
	}
	return out
}

func (h *RosterHandler) toRosterResponses(rosters []domain.RosterDetail) []rosterResponse {
	out := make([]rosterResponse, 0, len(rosters))
	for _, r := range rosters {
		out = append(out, h.toRosterResponse(r))
	}
	return out
}

func (h *RosterHandler) toRosterResponse(r domain.RosterDetail) rosterResponse {
	return rosterResponse{
		ID:         r.ID,
		Season:     r.Season,
		TeamID:     r.TeamID,
		PlayerID:   r.PlayerID,
		PositionID: r.PositionID,
		Number:     r.Number,
		Rating:     r.Rating,
		Experience: r.Experience,
		Player:     h.toPlayerResponse(r.Player),
		Position:   toPositionResponse(r.Position),
	}
}

func (h *RosterHandler) toRosterDetailResponse(p domain.RosterProfile) rosterDetailResponse {
	var starter *rosterStarterResponse
	if p.Starter != nil {
		starter = &rosterStarterResponse{Odflg: boolToInt(p.Starter.Odflg)}
	}

	career := make([]rosterCareerResponse, 0, len(p.Career))
	for _, rc := range p.Career {
		career = append(career, rosterCareerResponse{
			RosterID:     rc.RosterID,
			Season:       rc.Season,
			Team:         toTeamResponse(rc.Team),
			PositionName: rc.PositionName,
			Number:       rc.Number,
			Rating:       rc.Rating,
			Experience:   rc.Experience,
			IsStarter:    rc.IsStarter,
		})
	}

	return rosterDetailResponse{
		rosterResponse:   h.toRosterResponse(p.RosterDetail),
		Team:             toTeamResponse(p.Team),
		Starter:          starter,
		MaddenAttributes: toMaddenAttributeResponses(p.MaddenAttributes),
		Career:           career,
	}
}

func (h *RosterHandler) toPlayerResponse(p domain.Player) rosterPlayerResponse {
	return rosterPlayerResponse{
		ID:           p.ID,
		FirstName:    p.FirstName,
		LastName:     p.LastName,
		Birthday:     p.Birthday,
		Height:       p.HeightCm,
		Weight:       p.WeightKg,
		College:      p.College,
		DraftedTeam:  p.DraftedTeam,
		DraftedRound: p.DraftedRound,
		DraftedRank:  p.DraftedRank,
		DraftedYear:  p.DraftedYear,
		ImageFile:    p.ImageFile,
		BirthdayYear: birthdayYear(p.Birthday),
		BirthdayDate: birthdayDate(p.Birthday),
		ImageURL:     h.imageURL(p.ID, p.ImageFile),
	}
}

func toPositionResponse(p domain.PositionDetail) rosterPositionResponse {
	odflg := boolToInt(p.Odflg)
	abstractCategory := p.AbstractCategory
	category := p.Category
	return rosterPositionResponse{
		ID:               p.ID,
		AbstractCategory: &abstractCategory,
		Category:         &category,
		Name:             p.Name,
		Odflg:            &odflg,
	}
}

// boolToInt は攻守フラグなどの bool を 0/1 に変換する（Dart 側が int? 定義のため）。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// birthdayYear は "YYYYMMDD" 形式の birthday から年部分を取り出す。Laravel の Player モデルの
// getBirthdayYearAttribute アクセサに相当する。
func birthdayYear(birthday string) *int {
	if len(birthday) < 4 {
		return nil
	}
	y, err := strconv.Atoi(birthday[:4])
	if err != nil {
		return nil
	}
	return &y
}

// birthdayDate は "YYYYMMDD" 形式の birthday を "yyyy年MM月DD日" に整形する。Laravel の Player
// モデルの getBirthdayDateAttribute アクセサに相当する。
func birthdayDate(birthday string) *string {
	if len(birthday) < 8 {
		return nil
	}
	s := fmt.Sprintf("%s年%s月%s日", birthday[0:4], birthday[4:6], birthday[6:8])
	return &s
}

// imageURL は選手画像のURLを組み立てる。Laravel の Player モデルの getImageUrlAttribute
// アクセサに相当するが、画像配信先（AWS等）がまだ決まっていないため assetBaseURL が空なら
// image_file の有無に関わらず nil を返す。
func (h *RosterHandler) imageURL(playerID int32, imageFile *string) *string {
	if h.assetBaseURL == "" {
		return nil
	}
	if imageFile == nil || *imageFile == "" {
		url := h.assetBaseURL + "/no-image.jpg"
		return &url
	}
	url := fmt.Sprintf("%s/players/%d/%s", h.assetBaseURL, playerID, *imageFile)
	return &url
}

// toMaddenAttributeResponses は madden_attributes を key の昇順に並べた配列にする。
// Go の map は反復順が不定なので、レスポンスを安定させるために明示的にソートする。
func toMaddenAttributeResponses(attrs map[string]string) []maddenAttributeResponse {
	if len(attrs) == 0 {
		return []maddenAttributeResponse{}
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]maddenAttributeResponse, 0, len(keys))
	for _, k := range keys {
		label, ok := maddenimport.JapaneseLabel(k)
		if !ok {
			label = k
		}
		out = append(out, maddenAttributeResponse{Key: k, Label: label, Value: attrs[k]})
	}
	return out
}
