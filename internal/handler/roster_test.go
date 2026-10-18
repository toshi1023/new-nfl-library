package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

type stubRosterLister struct {
	rosters                 []domain.RosterDetail
	err                     error
	gotSeason, gotTeamID    int
	gotNames                []string
	gotSortBy, gotSortOrder string
}

func (s *stubRosterLister) Run(_ context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]domain.RosterDetail, error) {
	s.gotSeason, s.gotTeamID, s.gotNames = season, teamID, names
	s.gotSortBy, s.gotSortOrder = sortBy, sortOrder
	return s.rosters, s.err
}

type stubRosterDetailGetter struct {
	profile domain.RosterProfile
	err     error
	gotID   int
}

func (s *stubRosterDetailGetter) Run(_ context.Context, id int) (domain.RosterProfile, error) {
	s.gotID = id
	return s.profile, s.err
}

func newRosterRouter(lister RosterLister, assetBaseURL string) *gin.Engine {
	return newRosterRouterWithGetter(lister, &stubRosterDetailGetter{}, assetBaseURL)
}

func newRosterRouterWithGetter(lister RosterLister, getter RosterDetailGetter, assetBaseURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	RegisterValidatorTagName()
	r := gin.New()
	h := NewRosterHandler(lister, getter, assetBaseURL)
	r.GET("/api/v1/rosters", h.Index)
	r.GET("/api/v1/rosters/:id", h.Show)
	return r
}

func height(v float64) *float64 { return &v }
func str(v string) *string      { return &v }
func i32(v int32) *int32        { return &v }

func sampleRoster() domain.RosterDetail {
	return domain.RosterDetail{
		ID: 1, Season: 2024, TeamID: 1, PlayerID: 10, PositionID: 1,
		Number: i32(17), Rating: i32(92), Experience: i32(6),
		Player: domain.Player{
			ID: 10, FirstName: "Josh", LastName: "Allen", Birthday: "19960521",
			HeightCm: height(195.6), WeightKg: height(106.7), College: str("Wyoming"),
		},
		Position: domain.PositionDetail{ID: 1, AbstractCategory: 1, Category: 1, Name: "QB", Odflg: true},
	}
}

func sampleProfile() domain.RosterProfile {
	bills := domain.TeamDetail{ID: 1, City: "Buffalo", Name: "Bills", Conference: 1, Area: 1, ImageFile: str("bills.gif")}
	return domain.RosterProfile{
		RosterDetail: sampleRoster(),
		Team:         bills,
		Starter:      &domain.RosterStarter{Odflg: true},
		MaddenAttributes: map[string]string{
			"Speed":             "88",
			"___UNKNOWN_KEY___": "1",
		},
		Career: []domain.RosterCareer{
			{RosterID: 1, Season: 2024, Team: bills, PositionName: "QB", Number: i32(17), Rating: i32(92), Experience: i32(6), IsStarter: true},
			{RosterID: 2, Season: 2018, Team: bills, PositionName: "QB", Number: i32(17), IsStarter: false},
		},
	}
}

func TestRosterHandler_Index_Success(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2024, lister.gotSeason)
	assert.Equal(t, 1, lister.gotTeamID)

	body := w.Body.String()
	assert.Contains(t, body, `"message":"ロスター情報を取得しました"`)
	assert.Contains(t, body, `"firstname":"Josh"`)
	assert.Contains(t, body, `"birthday_year":1996`)
	assert.Contains(t, body, `"birthday_date":"1996年05月21日"`)
	assert.Contains(t, body, `"odflg":1`)
	assert.Contains(t, body, `"image_url":null`, "assetBaseURL が空なら image_url は null")
	assert.NotContains(t, body, `madden_attributes`, "一覧では madden_attributes を返さない（詳細 API でのみ返す）")
}

func TestRosterHandler_Index_EmptyResult(t *testing.T) {
	lister := &stubRosterLister{rosters: nil}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2012&team_id=1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"ロスター情報を取得しました","data":{"rosters":[]}}`, w.Body.String())
}

func TestRosterHandler_Index_NameFilter(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&name=Allen", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"Allen"}, lister.gotNames)
}

func TestRosterHandler_Index_NameOmitted(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, lister.gotNames)
}

func TestRosterHandler_Index_NameMultipleRepeatedParams(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&name=Josh&name=Diggs", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"Josh", "Diggs"}, lister.gotNames)
}

func TestRosterHandler_Index_NameCommaSeparated(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&name=Josh,+Diggs,Kincaid", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"Josh", "Diggs", "Kincaid"}, lister.gotNames)
}

func TestSplitNames(t *testing.T) {
	assert.Equal(t, []string{"Josh", "Allen"}, splitNames([]string{"Josh, Allen"}))
	assert.Equal(t, []string{"Josh", "Allen"}, splitNames([]string{"Josh", "Allen"}))
	assert.Equal(t, []string{"Josh", "Allen", "Diggs"}, splitNames([]string{"Josh,Allen", "Diggs"}))
	assert.Empty(t, splitNames(nil))
	assert.Empty(t, splitNames([]string{"", "  ", ","}))
}

func TestRosterHandler_Index_SortDefaults(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", lister.gotSortBy, "sort_by省略時はhandlerは空文字のままusecaseに渡す（既定値の適用はusecaseの責務）")
	assert.Equal(t, "", lister.gotSortOrder)
}

func TestRosterHandler_Index_SortByOrderGiven(t *testing.T) {
	lister := &stubRosterLister{rosters: []domain.RosterDetail{sampleRoster()}}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&sort_by=name&sort_order=asc", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "name", lister.gotSortBy)
	assert.Equal(t, "asc", lister.gotSortOrder)
}

func TestRosterHandler_Index_SortValidationError(t *testing.T) {
	lister := &stubRosterLister{}
	r := newRosterRouter(lister, "")

	t.Run("invalid sort_by", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&sort_by=bogus", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"sort_by":["oneof"]}}`, w.Body.String())
	})

	t.Run("invalid sort_order", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1&sort_order=bogus", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"sort_order":["oneof"]}}`, w.Body.String())
	})
}

func TestRosterHandler_Index_ImageURL(t *testing.T) {
	roster := sampleRoster()
	lister := &stubRosterLister{rosters: []domain.RosterDetail{roster}}
	r := newRosterRouter(lister, "https://assets.example.com")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1", nil))

	assert.Contains(t, w.Body.String(), `"image_url":"https://assets.example.com/no-image.jpg"`)
}

func TestRosterHandler_Index_ValidationError(t *testing.T) {
	lister := &stubRosterLister{}
	r := newRosterRouter(lister, "")

	t.Run("season missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?team_id=1", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"season":["required"]}}`, w.Body.String())
	})

	t.Run("season too old", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2000&team_id=1", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"season":["min"]}}`, w.Body.String())
	})

	t.Run("team_id out of range", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=99", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"team_id":["max"]}}`, w.Body.String())
	})
}

func TestRosterHandler_Index_UsecaseError(t *testing.T) {
	lister := &stubRosterLister{err: errors.New("db down")}
	r := newRosterRouter(lister, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters?season=2024&team_id=1", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message":"サーバー内部でエラーが発生しました","code":"E0006"}`, w.Body.String())
}

func TestRosterHandler_Show_Success(t *testing.T) {
	getter := &stubRosterDetailGetter{profile: sampleProfile()}
	r := newRosterRouterWithGetter(&stubRosterLister{}, getter, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, getter.gotID)

	var got struct {
		Message string `json:"message"`
		Data    struct {
			Roster map[string]json.RawMessage `json:"roster"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "ロスター詳細を取得しました", got.Message)

	roster := got.Data.Roster
	assert.JSONEq(t, `1`, string(roster["id"]), "一覧と同じ項目がフラットに含まれる")
	assert.Contains(t, string(roster["player"]), `"birthday_date":"1996年05月21日"`)
	assert.JSONEq(t, `{"id":1,"city":"Buffalo","name":"Bills","conference":"AFC","area":"東","image_file":"bills.gif","back_image_file":null}`, string(roster["team"]))
	assert.JSONEq(t, `{"odflg":1}`, string(roster["starter"]))
	assert.JSONEq(t, `[
		{"key":"Speed","label":"スピード","value":"88"},
		{"key":"___UNKNOWN_KEY___","label":"___UNKNOWN_KEY___","value":"1"}
	]`, string(roster["madden_attributes"]), "未知のキーはラベルが無ければキーそのものを使う")
	assert.JSONEq(t, `[
		{"roster_id":1,"season":2024,"team":{"id":1,"city":"Buffalo","name":"Bills","conference":"AFC","area":"東","image_file":"bills.gif","back_image_file":null},
		 "position_name":"QB","number":17,"rating":92,"experience":6,"is_starter":true},
		{"roster_id":2,"season":2018,"team":{"id":1,"city":"Buffalo","name":"Bills","conference":"AFC","area":"東","image_file":"bills.gif","back_image_file":null},
		 "position_name":"QB","number":17,"rating":null,"experience":null,"is_starter":false}
	]`, string(roster["career"]))
}

func TestRosterHandler_Show_NotStarterNoMadden(t *testing.T) {
	profile := sampleProfile()
	profile.Starter = nil
	profile.MaddenAttributes = nil
	profile.Career = nil
	r := newRosterRouterWithGetter(&stubRosterLister{}, &stubRosterDetailGetter{profile: profile}, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/1", nil))

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"starter":null`)
	assert.Contains(t, body, `"madden_attributes":[]`)
	assert.Contains(t, body, `"career":[]`)
}

func TestRosterHandler_Show_NotFound(t *testing.T) {
	getter := &stubRosterDetailGetter{err: fmt.Errorf("roster id 999: %w", domain.ErrNotFound)}
	r := newRosterRouterWithGetter(&stubRosterLister{}, getter, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/999", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"message":"roster id 999: not found","code":"E0004"}`, w.Body.String())
}

func TestRosterHandler_Show_ValidationError(t *testing.T) {
	getter := &stubRosterDetailGetter{}
	r := newRosterRouterWithGetter(&stubRosterLister{}, getter, "")

	t.Run("id is zero", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/0", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"id":["required"]}}`, w.Body.String())
	})

	t.Run("id is negative", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/-1", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"id":["min"]}}`, w.Body.String())
	})

	t.Run("id is not a number", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/abc", nil))

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"E0001"`)
	})

	assert.Zero(t, getter.gotID, "不正な id では usecase を呼ばない")
}

func TestRosterHandler_Show_UsecaseError(t *testing.T) {
	getter := &stubRosterDetailGetter{err: errors.New("db down")}
	r := newRosterRouterWithGetter(&stubRosterLister{}, getter, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/rosters/1", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message":"サーバー内部でエラーが発生しました","code":"E0006"}`, w.Body.String())
}
