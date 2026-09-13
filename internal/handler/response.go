// Package handler は Gin ハンドラを提供する。HTTP とusecase の変換のみを行い、ビジネスロジックは書かない。
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// エラーコード。api/openapi.yaml の ErrorDetail.code と一致させる。
const (
	CodeValidation         = "VALIDATION_ERROR"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeInternal           = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// ErrorResponse は統一エラーレスポンス。
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail はエラーの詳細。
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError はエラーを HTTP ステータスと統一 JSON に変換して返す。
// ドメインエラー → HTTP ステータスの変換はここでのみ行う。
func respondError(c *gin.Context, err error) {
	status, code, msg := classify(err)

	if status >= http.StatusInternalServerError {
		logging.FromContext(c.Request.Context()).Error("request failed", slog.Any("err", err))
	}

	c.AbortWithStatusJSON(status, ErrorResponse{Error: ErrorDetail{Code: code, Message: msg}})
}

// respondBindError はリクエストのバインド / バリデーション失敗を 400 に変換する。
func respondBindError(c *gin.Context, err error) {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		respondError(c, domain.ErrValidation)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{Code: CodeValidation, Message: "invalid request body"},
	})
}

func classify(err error) (status int, code, msg string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, CodeValidation, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, CodeUnauthorized, err.Error()
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, CodeForbidden, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, CodeNotFound, err.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, CodeConflict, err.Error()
	default:
		// 内部エラーの詳細はクライアントに出さない。
		return http.StatusInternalServerError, CodeInternal, "internal server error"
	}
}
