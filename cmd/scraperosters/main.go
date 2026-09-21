// Command scraperosters は pro-football-reference.com から NFL の全チームロスターをスクレイプし、CSV に一括出力する。
//
// 使い方:
//
//	go run ./cmd/scraperosters -season 2024                 # 全 32 チーム → outputs/csv/2024_rosters.csv
//	go run ./cmd/scraperosters -season 2024 -teams buf,nwe  # チームを絞る
//	go run ./cmd/scraperosters -season 2024 -out buf.csv -delay 5s
//
// CSV は常に outputs/csv/ 配下に出力する。-out はその中でのファイル名。
//
// 環境変数:
//
//	CHROME_WS_URL  リモート Chrome の DevTools URL（例: ws://localhost:9222）。未設定ならローカルの Chrome を起動する
//	LOG_LEVEL      debug / info / warn / error（既定 info）
//
// 注意: スクレイプ先は Cloudflare の bot 対策を使っており、headless Chrome では通過できない。
// 既定では画面ありの Chrome が起動する。実行中はウィンドウを閉じないこと。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/infra/csvwriter"
	"github.com/toshi1023/new-nfl-library/internal/infra/pfr"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/usecase"
)

const defaultSeason = 2020

// outputDir は CSV の出力先ディレクトリ（プロジェクトルートからの相対パス）。無ければ実行時に作成される。
var outputDir = filepath.Join("outputs", "csv")

func main() {
	if err := run(); err != nil {
		slog.Error("scrape rosters failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	// ---- フラグ(フラグ名, デフォルト値, 説明) ※例：go run ./cmd/scraperosters -season 2024 -teams buf,nwe -delay 5s ----
	var (
		season   = flag.Int("season", defaultSeason, "対象シーズン（2012 以降）")
		teamsArg = flag.String("teams", "", "対象チーム略称のカンマ区切り（例: buf,nwe）。省略時は全 32 チーム")
		out      = flag.String("out", "", "出力 CSV のファイル名。outputs/csv/ 配下に置かれる（既定: <season>_rosters.csv）")
		delay    = flag.Duration("delay", 3*time.Second, "チーム間の待機時間")
		timeout  = flag.Duration("timeout", 30*time.Second, "1 ページの取得タイムアウト")
		headless = flag.Bool("headless", false, "Chrome を headless で起動する（スクレイプ先の bot 対策により headless では取得できないため通常は false）")
		dumpDir  = flag.String("dump-dir", "", "取得した HTML を保存するディレクトリ（デバッグ用）")
	)
	// コマンドの引数で渡された値をフラグにセットする。フラグの既定値は上記の変数初期化で設定済み。
	flag.Parse()

	// ---- ロガー ----
	level := slog.LevelInfo
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if err := level.UnmarshalText([]byte(v)); err != nil {
			return fmt.Errorf("invalid LOG_LEVEL %q: %w", v, err)
		}
	}
	logger := logging.New(os.Stdout, level)
	slog.SetDefault(logger)

	// ---- 対象チーム ----
	teams, err := resolveTeams(*teamsArg)
	if err != nil {
		return err
	}
	outName := *out
	if outName == "" {
		outName = fmt.Sprintf("%d_rosters.csv", *season)
	}
	outPath := filepath.Join(outputDir, outName)

	// Ctrl+C / SIGTERM で中断できるようにする。中断時もそれまでの取得分は CSV に残る
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx = logging.WithLogger(ctx, logger)

	// ---- 依存の組み立て ----
	fetcher, err := pfr.NewFetcher(ctx, pfr.Options{
		RemoteURL:   os.Getenv("CHROME_WS_URL"),
		Headless:    *headless,
		PageTimeout: *timeout,
		DumpDir:     *dumpDir,
	})
	if err != nil {
		return err
	}
	defer fetcher.Close()

	writer := csvwriter.NewWriter(outPath)
	uc := usecase.NewScrapeRosters(fetcher, writer, *delay)

	// ---- 実行 ----
	logger.Info("scrape rosters start",
		slog.Int("season", *season),
		slog.Int("teams", len(teams)),
		slog.String("out", outPath),
	)
	n, err := uc.Run(ctx, *season, teams)
	if n > 0 {
		logger.Info("csv written", slog.String("path", outPath), slog.Int("rows", n))
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Warn("interrupted; partial result written", slog.Int("rows", n))
		}
		return err
	}
	logger.Info("scrape rosters all successfully", slog.Int("rows", n))
	return nil
}

// resolveTeams はカンマ区切りの略称を Team に変換する。空なら全チーム。
func resolveTeams(arg string) ([]domain.Team, error) {
	// 引数でチームの指定をしていない場合は全チームを返す。
	if strings.TrimSpace(arg) == "" {
		return domain.AllTeams(), nil
	}
	var teams []domain.Team
	for _, abbr := range strings.Split(arg, ",") {
		abbr = strings.ToLower(strings.TrimSpace(abbr))
		if abbr == "" {
			continue
		}
		t, err := domain.FindTeam(abbr)
		if err != nil {
			return nil, fmt.Errorf("unknown team %q: %w", abbr, err)
		}
		teams = append(teams, t)
	}
	return teams, nil
}
