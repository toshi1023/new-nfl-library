package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubPinger struct{ err error }

func (s stubPinger) Ping(context.Context) error { return s.err }

func newHealthRouter(p Pinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHealthHandler(p)
	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)
	return r
}

func TestHealthHandler_Healthz(t *testing.T) {
	r := newHealthRouter(stubPinger{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestHealthHandler_Readyz(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		r := newHealthRouter(stubPinger{})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
	})

	t.Run("not ready", func(t *testing.T) {
		r := newHealthRouter(stubPinger{err: errors.New("db down")})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil))

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.JSONEq(t, `{"message":"依存先に接続できません","code":"E0007"}`, w.Body.String())
	})
}
