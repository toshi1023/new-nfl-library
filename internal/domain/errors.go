// Package domain はエンティティ・値オブジェクト・リポジトリインターフェース・ドメインエラーを定義する。
// 他の internal パッケージを import してはならない。
package domain

import "errors"

// センチネルエラー。usecase / repository は %w でこれらをラップして返し、
// HTTP ステータスへの変換は handler 層でのみ行う。
var (
	// ErrNotFound はリソースが存在しない。
	ErrNotFound = errors.New("not found")
	// ErrConflict はリソースの一意性制約などに違反した。
	ErrConflict = errors.New("conflict")
	// ErrUnauthorized は認証されていない。
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden は認証済みだが権限がない。
	ErrForbidden = errors.New("forbidden")
	// ErrValidation は入力値が不正。
	ErrValidation = errors.New("validation error")
)
