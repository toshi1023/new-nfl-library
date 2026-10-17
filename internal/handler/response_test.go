package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"validation", fmt.Errorf("age: %w", domain.ErrValidation), http.StatusUnprocessableEntity, CodeValidation},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, CodeUnauthorized},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, CodeForbidden},
		{"not found", fmt.Errorf("team: %w", domain.ErrNotFound), http.StatusNotFound, CodeNotFound},
		{"conflict", domain.ErrConflict, http.StatusConflict, CodeConflict},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, msg := classify(tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantCode, code)
			if tt.wantStatus == http.StatusInternalServerError {
				assert.Equal(t, "サーバー内部でエラーが発生しました", msg, "内部エラーの詳細を漏らさない")
			}
		})
	}
}

func TestRespondSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		respondSuccess(c, "ok", gin.H{"rosters": []int{1, 2}})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"ok","data":{"rosters":[1,2]}}`, w.Body.String())
}

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		respondError(c, fmt.Errorf("team: %w", domain.ErrNotFound))
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"message":"team: not found","code":"E0004"}`, w.Body.String())
}

func TestRespondValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		respondValidationError(c, "入力値が不正です", map[string][]string{"season": {"2012以降の値を指定してください"}})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.JSONEq(t, `{"message":"入力値が不正です","code":"E0001","errors":{"season":["2012以降の値を指定してください"]}}`, w.Body.String())
}

func TestRespondBindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type req struct {
		Name string `json:"name" binding:"required"`
	}
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		var in req
		if err := c.ShouldBindJSON(&in); err != nil {
			respondBindError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	t.Run("validation tag failure", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{}`))
		httpReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"E0001"`)
	})

	t.Run("malformed json", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{`))
		httpReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"E0001"`)
	})

	t.Run("ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"name":"x"}`))
		httpReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
