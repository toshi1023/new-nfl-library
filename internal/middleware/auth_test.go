package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAuthRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", Auth(secret), func(c *gin.Context) {
		sub, _ := SubjectFromContext(c.Request.Context())
		c.String(http.StatusOK, sub)
	})
	return r
}

func signToken(t *testing.T, secret, sub string, exp time.Time) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(exp),
	})
	s, err := tok.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func TestAuth(t *testing.T) {
	const secret = "test-secret"
	r := newAuthRouter(secret)

	t.Run("valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+signToken(t, secret, "user-1", time.Now().Add(time.Hour)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-1", w.Body.String())
	})

	t.Run("missing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":{"code":"UNAUTHORIZED","message":"missing bearer token"}}`, w.Body.String())
	})

	t.Run("wrong secret", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+signToken(t, "other", "user-1", time.Now().Add(time.Hour)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+signToken(t, secret, "user-1", time.Now().Add(-time.Hour)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
