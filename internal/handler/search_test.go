package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

type stubSeasonLister struct {
	seasons []int32
	err     error
}

func (s *stubSeasonLister) Run(_ context.Context) ([]int32, error) {
	return s.seasons, s.err
}

type stubTeamLister struct {
	teams []domain.TeamDetail
	err   error
}

func (s *stubTeamLister) Run(_ context.Context) ([]domain.TeamDetail, error) {
	return s.teams, s.err
}

func newSearchRouter(seasons SeasonLister, teams TeamLister) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSearchHandler(seasons, teams)
	r.GET("/api/v1/search/seasons", h.Seasons)
	r.GET("/api/v1/search/teams", h.Teams)
	return r
}

func TestSearchHandler_Seasons_Success(t *testing.T) {
	seasons := &stubSeasonLister{seasons: []int32{2025, 2024, 2023}}
	r := newSearchRouter(seasons, &stubTeamLister{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/seasons", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"シーズン一覧を取得しました","data":{"seasons":[{"season":2025},{"season":2024},{"season":2023}]}}`, w.Body.String())
}

func TestSearchHandler_Seasons_Empty(t *testing.T) {
	r := newSearchRouter(&stubSeasonLister{}, &stubTeamLister{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/seasons", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"シーズン一覧を取得しました","data":{"seasons":[]}}`, w.Body.String())
}

func TestSearchHandler_Seasons_UsecaseError(t *testing.T) {
	r := newSearchRouter(&stubSeasonLister{err: errors.New("db down")}, &stubTeamLister{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/seasons", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message":"サーバー内部でエラーが発生しました","code":"E0006"}`, w.Body.String())
}

func TestSearchHandler_Teams_Success(t *testing.T) {
	teams := &stubTeamLister{teams: []domain.TeamDetail{
		{ID: 1, City: "Buffalo", Name: "Bills", Conference: 1, Area: 1, ImageFile: str("bills.gif"), BackImageFile: str("bills_cover.jpg")},
		{ID: 17, City: "Dallas", Name: "Cowboys", Conference: 2, Area: 1},
		{ID: 99, City: "Unknown", Name: "Team", Conference: 9, Area: 9},
	}}
	r := newSearchRouter(&stubSeasonLister{}, teams)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/teams", nil))

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"message":"チーム一覧を取得しました"`)
	assert.Contains(t, body, `{"id":1,"city":"Buffalo","name":"Bills","conference":"AFC","area":"東","image_file":"bills.gif","back_image_file":"bills_cover.jpg"}`)
	assert.Contains(t, body, `{"id":17,"city":"Dallas","name":"Cowboys","conference":"NFC","area":"東","image_file":null,"back_image_file":null}`)
	assert.Contains(t, body, `"conference":"9","area":"9"`, "未知のコードは数値を文字列化してフォールバックする")
}

func TestSearchHandler_Teams_Empty(t *testing.T) {
	r := newSearchRouter(&stubSeasonLister{}, &stubTeamLister{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/teams", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"チーム一覧を取得しました","data":{"teams":[]}}`, w.Body.String())
}

func TestSearchHandler_Teams_UsecaseError(t *testing.T) {
	r := newSearchRouter(&stubSeasonLister{}, &stubTeamLister{err: errors.New("db down")})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search/teams", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message":"サーバー内部でエラーが発生しました","code":"E0006"}`, w.Body.String())
}
