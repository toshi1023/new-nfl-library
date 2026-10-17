// Package router は全ルート定義を 1 箇所に集約する。
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/handler"
	"github.com/toshi1023/new-nfl-library/internal/middleware"
)

// Deps はルーティングに必要な依存。
type Deps struct {
	Logger    *slog.Logger
	JWTSecret string
	Health    *handler.HealthHandler
	Roster    *handler.RosterHandler
}

// New は Gin エンジンを生成し、ミドルウェアとルートを登録する。
func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(d.Logger),
		middleware.Recovery(),
	)

	r.GET("/healthz", d.Health.Healthz)
	r.GET("/readyz", d.Health.Readyz)

	v1 := r.Group("/api/v1")

	v1.GET("/rosters", d.Roster.Index)

	// 認証不要のルートは v1 に、認証必須のルートは authed に登録する。
	authed := v1.Group("", middleware.Auth(d.JWTSecret))
	_ = authed

	return r
}
