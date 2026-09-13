package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// Recovery は panic を捕捉してログに記録し、統一形式の 500 レスポンスを返す。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logging.FromContext(c.Request.Context()).Error("panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "internal server error",
					},
				})
			}
		}()
		c.Next()
	}
}
