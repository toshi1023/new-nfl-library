// Command seed はマスタデータ（teams / positions / formations / pf_relations / users / settings / foul_rules）を投入する。
// Laravel の `php artisan db:seed` に相当する。
//
// 使い方:
//
//	go run ./cmd/seed          # データが無いテーブルにだけ投入する
//	go run ./cmd/seed -force   # 対象テーブルを TRUNCATE してから投入し直す
//
// 環境変数:
//
//	DATABASE_URL        接続文字列（必須）
//	SEED_ROOT_EMAIL     root ユーザーのメールアドレス
//	SEED_ROOT_PASSWORD  root ユーザーのパスワード（bcrypt でハッシュして保存）
//	LOG_LEVEL           debug / info / warn / error（既定 info）
//
// SEED_ROOT_EMAIL / SEED_ROOT_PASSWORD が未設定なら users と settings はスキップする。
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/seed"
)

type config struct {
	DatabaseURL  string `envconfig:"DATABASE_URL" required:"true"`
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	RootEmail    string `envconfig:"SEED_ROOT_EMAIL"`
	RootPassword string `envconfig:"SEED_ROOT_PASSWORD"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	force := flag.Bool("force", false, "対象テーブルを TRUNCATE してから投入し直す")
	flag.Parse()

	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("invalid LOG_LEVEL %q: %w", cfg.LogLevel, err)
	}
	logger := logging.New(os.Stdout, level)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx = logging.WithLogger(ctx, logger)

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := seed.New(pool).Run(ctx, seed.Options{
		Force:        *force,
		RootEmail:    cfg.RootEmail,
		RootPassword: cfg.RootPassword,
	}); err != nil {
		return err
	}
	logger.Info("seed completed")
	return nil
}
