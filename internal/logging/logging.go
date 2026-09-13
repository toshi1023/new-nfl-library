// Package logging は slog ロガーの生成と context への格納・取り出しを提供する。
package logging

import (
	"context"
	"io"
	"log/slog"
)

type ctxKey struct{}

// New は JSON 形式の構造化ロガーを生成する。
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

// WithLogger は ctx にロガーを格納した新しい ctx を返す。
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext は ctx に格納されたロガーを返す。未格納なら slog.Default() を返す。
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
