// Package handler は Gin ハンドラを提供する。HTTP とusecase の変換のみを行い、ビジネスロジックは書かない。
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// RegisterValidatorTagName は Gin のバリデーションエラーのフィールド名を、Go の構造体フィールド名
// （例: "TeamID"）ではなく form/uri/json タグ（例: "team_id"。クライアントが実際に送るキー）にする。
// main 起動時に一度だけ呼ぶこと（グローバルな副作用があるため init() では行わない）。
func RegisterValidatorTagName() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("form")
		if name == "" {
			name = fld.Tag.Get("uri")
		}
		if name == "" {
			name = fld.Tag.Get("json")
		}
		name, _, _ = strings.Cut(name, ",")
		if name == "-" {
			return ""
		}
		return name
	})
}

// エラーコード。HTTP ステータスとは別軸のアプリ独自コード。api/openapi.yaml の ErrorCode と一致させる。
// Flutter 側（nfl_library_app）の ApiResult.error / ApiResult.validationError が参照する。
const (
	CodeValidation   = "E0001" // 入力値が不正
	CodeUnauthorized = "E0002" // 認証されていない
	CodeForbidden    = "E0003" // 権限がない
	CodeNotFound     = "E0004" // リソースが見つからない
	CodeConflict     = "E0005" // リソースが競合している
	CodeInternal     = "E0006" // サーバー内部エラー
	CodeUnavailable  = "E0007" // 依存先サービスが利用不可
)

// SuccessResponse は成功時の統一レスポンス。
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// ErrorResponse はエラー時の統一レスポンス。
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationErrorResponse はバリデーションエラー（422）時の統一レスポンス。
type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Code    string              `json:"code"`
	Errors  map[string][]string `json:"errors"`
}

// respondSuccess は 200 として message と data を返す。
func respondSuccess(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, SuccessResponse{Message: message, Data: data})
}

// respondError はエラーを HTTP ステータスと統一 JSON に変換して返す。
// ドメインエラー → HTTP ステータス・コードの変換はここでのみ行う。
func respondError(c *gin.Context, err error) {
	status, code, msg := classify(err)

	if status >= http.StatusInternalServerError {
		logging.FromContext(c.Request.Context()).Error("request failed", slog.Any("err", err))
	}

	c.AbortWithStatusJSON(status, ErrorResponse{Message: msg, Code: code})
}

// respondValidationError は 422 としてフィールド別のエラーメッセージを返す。
func respondValidationError(c *gin.Context, message string, errs map[string][]string) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Message: message,
		Code:    CodeValidation,
		Errors:  errs,
	})
}

// respondBindError はリクエストのバインド / バリデーション失敗を 422 に変換する。
func respondBindError(c *gin.Context, err error) {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		errs := make(map[string][]string, len(verrs))
		for _, fe := range verrs {
			field := fe.Field()
			errs[field] = append(errs[field], fe.Tag())
		}
		respondValidationError(c, "入力値が不正です", errs)
		return
	}
	respondValidationError(c, "入力値が不正です", map[string][]string{"_": {"invalid request"}})
}

func classify(err error) (status int, code, msg string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, CodeValidation, err.Error()
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
		return http.StatusInternalServerError, CodeInternal, "サーバー内部でエラーが発生しました"
	}
}
