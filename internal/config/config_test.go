package config

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetenv はテスト中だけ環境変数を未設定状態にする（テスト終了時に元の値へ戻す）。
func unsetenv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		t.Setenv(k, "") // クリーンアップ登録のため
		require.NoError(t, os.Unsetenv(k))
	}
}

func TestLoad(t *testing.T) {
	t.Run("defaults in development", func(t *testing.T) {
		unsetenv(t, "APP_ENV", "APP_PORT", "JWT_SECRET", "LOG_LEVEL")
		t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/app")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, EnvDevelopment, cfg.AppEnv)
		assert.Equal(t, ":8080", cfg.Addr())
		assert.True(t, cfg.IsDevelopment())

		lvl, err := cfg.SlogLevel()
		require.NoError(t, err)
		assert.Equal(t, slog.LevelInfo, lvl)
	})

	t.Run("DATABASE_URL is required", func(t *testing.T) {
		unsetenv(t, "DATABASE_URL")
		_, err := Load()
		require.Error(t, err)
	})

	t.Run("DATABASE_URL must not be empty", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		_, err := Load()
		require.Error(t, err)
	})

	t.Run("JWT_SECRET is required outside development", func(t *testing.T) {
		unsetenv(t, "JWT_SECRET")
		t.Setenv("APP_ENV", EnvProduction)
		t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/app")
		_, err := Load()
		require.Error(t, err)
	})

	t.Run("invalid LOG_LEVEL", func(t *testing.T) {
		t.Setenv("APP_ENV", EnvDevelopment)
		t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/app")
		t.Setenv("LOG_LEVEL", "verbose")
		_, err := Load()
		require.Error(t, err)
	})
}
