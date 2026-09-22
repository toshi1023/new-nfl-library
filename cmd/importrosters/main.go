// Command importrosters は cmd/scraperosters が出力した CSV（outputs/csv/*.csv）を読み込み、
// players / rosters テーブルへ反映する。Laravel の ScrapeRoster（DB 書き込み部分）に相当する。
//
// 使い方:
//
//	go run ./cmd/importrosters -csv outputs/csv/2024_rosters.csv
//	go run ./cmd/importrosters -csv outputs/csv/2024_rosters.csv -strict
//
// 事前に `make seed` で teams / positions が投入されている必要がある。
//
// 環境変数:
//
//	DATABASE_URL  接続文字列（必須）
//	LOG_LEVEL     debug / info / warn / error（既定 info）
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

	"github.com/toshi1023/new-nfl-library/internal/infra/csvwriter"
	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/rosterimport"
)

type config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("import rosters failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	csvPath := flag.String("csv", "", "読み込む CSV パス（必須。cmd/scraperosters の出力形式）")
	strict := flag.Bool("strict", false, "positions マスタに無いポジションが出た時点で処理全体を中断する（既定はスキップして続行）")
	flag.Parse()

	if *csvPath == "" {
		return fmt.Errorf("-csv is required")
	}

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

	entries, err := csvwriter.ReadRosterEntriesFile(*csvPath)
	if err != nil {
		return err
	}
	logger.Info("csv loaded", slog.String("path", *csvPath), slog.Int("rows", len(entries)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx = logging.WithLogger(ctx, logger)

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	result, err := rosterimport.New(pool).Run(ctx, entries, rosterimport.Options{Strict: *strict})
	if err != nil {
		return err
	}
	logger.Info("import completed",
		slog.Int("players_created", result.PlayersCreated),
		slog.Int("players_reused", result.PlayersReused),
		slog.Int("rosters_upserted", result.RostersUpserted),
		slog.Int("skipped", result.Skipped),
	)
	return nil
}
