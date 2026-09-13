package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// Pinger は依存先（DB など）への到達性を確認できるもの。
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler は liveness / readiness エンドポイントを提供する。
type HealthHandler struct {
	pinger Pinger
}

// NewHealthHandler は HealthHandler を生成する。
func NewHealthHandler(pinger Pinger) *HealthHandler {
	return &HealthHandler{pinger: pinger}
}

type healthResponse struct {
	Status string `json:"status"`
}

// Healthz はプロセスが稼働していれば 200 を返す。
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

// Readyz は依存先に接続できれば 200、できなければ 503 を返す。
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		logging.FromContext(c.Request.Context()).Warn("readiness check failed", slog.Any("err", err))
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: ErrorDetail{Code: CodeServiceUnavailable, Message: "dependency unavailable"},
		})
		return
	}
	c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}
