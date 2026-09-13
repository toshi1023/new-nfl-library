// Package middleware は Gin のミドルウェア（リクエスト ID・ログ・リカバリ・認証）を提供する。
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HeaderRequestID はリクエスト ID を受け渡す HTTP ヘッダ名。
const HeaderRequestID = "X-Request-ID"

type requestIDKey struct{}

// RequestID はリクエストごとに ID を発行（またはヘッダから引き継ぎ）し、ctx とレスポンスヘッダに設定する。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Header(HeaderRequestID, id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey{}, id))
		c.Next()
	}
}

// RequestIDFromContext は ctx からリクエスト ID を取り出す。未設定なら空文字。
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
