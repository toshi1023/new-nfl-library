package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// Logger はリクエスト ID を付与したロガーを ctx に格納し、リクエスト完了時にアクセスログを出力する。
// RequestID ミドルウェアの後に登録すること。
func Logger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		l := base.With(slog.String("request_id", RequestIDFromContext(ctx)))
		c.Request = c.Request.WithContext(logging.WithLogger(ctx, l))

		c.Next()

		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("bytes", c.Writer.Size()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		switch {
		case c.Writer.Status() >= 500:
			l.Error("request", attrs...)
		case c.Writer.Status() >= 400:
			l.Warn("request", attrs...)
		default:
			l.Info("request", attrs...)
		}
	}
}
