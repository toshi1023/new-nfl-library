package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type subjectKey struct{}

// Auth は Authorization: Bearer <JWT> を検証し、subject を ctx に格納する。
// 署名は HS256 のみ許可する。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			abortUnauthorized(c, "missing bearer token")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, prefix))

		claims := jwt.RegisteredClaims{}
		_, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || claims.Subject == "" {
			abortUnauthorized(c, "invalid token")
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), subjectKey{}, claims.Subject))
		c.Next()
	}
}

// SubjectFromContext は Auth ミドルウェアが格納した subject（ユーザー識別子）を返す。
func SubjectFromContext(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(subjectKey{}).(string)
	return s, ok && s != ""
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.Header("WWW-Authenticate", `Bearer realm="api"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": msg,
		},
	})
}
