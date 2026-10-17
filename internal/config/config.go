// Package config は環境変数からアプリケーション設定を読み込み、検証する。
package config

import (
	"fmt"
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

// 環境名の定数。
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Config はアプリケーション全体の設定。
type Config struct {
	AppEnv      string `envconfig:"APP_ENV" default:"development"`
	AppPort     int    `envconfig:"APP_PORT" default:"8080"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	// AssetBaseURL は選手画像の配信元URL。未設定（既定）なら image_url は常に null になる。
	AssetBaseURL string `envconfig:"ASSET_BASE_URL" default:""`
}

// Load は環境変数から Config を生成し、値を検証する。
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("process env: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	switch c.AppEnv {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("APP_ENV must be one of development/staging/production, got %q", c.AppEnv)
	}
	if c.AppPort <= 0 || c.AppPort > 65535 {
		return fmt.Errorf("APP_PORT must be 1-65535, got %d", c.AppPort)
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.AppEnv != EnvDevelopment && c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required when APP_ENV=%s", c.AppEnv)
	}
	if _, err := c.SlogLevel(); err != nil {
		return err
	}
	return nil
}

// IsDevelopment は開発環境かどうかを返す。
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == EnvDevelopment
}

// Addr は HTTP サーバーのリッスンアドレスを返す。
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

// SlogLevel は LOG_LEVEL を slog.Level に変換する。
func (c *Config) SlogLevel() (slog.Level, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(c.LogLevel)); err != nil {
		return 0, fmt.Errorf("LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel)
	}
	return lvl, nil
}
