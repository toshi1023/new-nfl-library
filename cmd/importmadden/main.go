// Command importmadden は Madden NFL の評価値CSVを読み込み、rosters テーブルの position_id / rating /
// number / madden_attributes を更新する。Laravel の CsvImport コマンドに相当する。
//
// 使い方:
//
//	go run ./cmd/importmadden                               # 全シーズン一括（db/data/madden 配下）
//	go run ./cmd/importmadden -season 2024 -dir path/to/2024
//	go run ./cmd/importmadden -season 2012 -csv a.csv -csv b.csv
//	go run ./cmd/importmadden -season 2024 -dir path/to/2024 -strict
//
// -season を省略すると全シーズン一括モードになる。-dir（既定 db/data/madden）直下の、名前が
// 4桁の年であるサブディレクトリをそれぞれ1シーズンとして扱い、順に処理する
// （db/data/madden/2024/ のような構成。年ごとのサブディレクトリの中身はチーム別の複数ファイルでも、
// 全チームがまとまった単一ファイルでもよい）。この場合 -csv は使えない（どのシーズンの
// ファイルか判断できないため）。
//
// -season を指定すると、その1シーズンだけを対象にする従来の挙動になる。-dir はそのディレクトリ
// 直下の *.csv をすべて対象にする（サブディレクトリは見ない）。-dir と -csv は併用できる。
//
// どちらのモードでも、CSV に Team 列が無い年（2014・2015年分）はファイル名の先頭トークン
// （"_" 区切り）をチーム略称とみなして解決する（例: atl_madden_nfl_15.csv → atl）。
//
// 事前に `make seed` で teams / positions が、`cmd/importrosters` で対象シーズンの roster が
// 投入済みである必要がある（該当する roster が無い行はスキップする）。
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

	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/maddenimport"
)

// defaultSeasonsDir は -season を省略した全シーズン一括モードで使うルートディレクトリの既定値。
const defaultSeasonsDir = "db/data/madden"

type config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// csvFlags は -csv を複数回指定できるようにする flag.Value 実装。
type csvFlags []string

func (f *csvFlags) String() string { return fmt.Sprint([]string(*f)) }
func (f *csvFlags) Set(v string) error {
	*f = append(*f, v)
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("import madden ratings failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	season := flag.Int("season", 0, "対象シーズン。省略すると -dir 配下の全シーズン（年名のサブディレクトリ）を一括処理する")
	dir := flag.String("dir", "", "単一シーズンモード: このディレクトリ直下の *.csv をすべて対象にする（既定なし）。\n全シーズンモード: 年ごとのサブディレクトリを持つルートディレクトリ（既定 db/data/madden）")
	strict := flag.Bool("strict", false, "positions マスタに無いポジションが出た時点で処理全体を中断する（既定はスキップして続行）")
	var csvPaths csvFlags
	flag.Var(&csvPaths, "csv", "CSVパス（複数指定可、単一シーズンモードのみ）")
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
	im := maddenimport.New(pool)

	if *season == 0 {
		return runAllSeasons(ctx, im, *dir, csvPaths, *strict, logger)
	}
	return runSingleSeason(ctx, im, *season, *dir, csvPaths, *strict, logger)
}

// runAllSeasons は -season 省略時の全シーズン一括モード。
func runAllSeasons(ctx context.Context, im *maddenimport.Importer, dir string, csvPaths csvFlags, strict bool, logger *slog.Logger) error {
	if len(csvPaths) > 0 {
		return fmt.Errorf("-csv は -season と併用する必要がある（全シーズンモードではどのシーズンのファイルか判断できない）")
	}
	if dir == "" {
		dir = defaultSeasonsDir
	}
	_, err := im.RunSeasons(ctx, dir, 0, maddenimport.Options{Strict: strict}, logger)
	return err
}

// runSingleSeason は -season 指定時の従来の挙動。
func runSingleSeason(ctx context.Context, im *maddenimport.Importer, season int, dir string, csvPaths csvFlags, strict bool, logger *slog.Logger) error {
	paths, err := maddenimport.CollectPaths(dir, csvPaths)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no CSV files given: specify -csv and/or -dir")
	}

	files, err := maddenimport.LoadFiles(paths, logger)
	if err != nil {
		return err
	}
	totalRows := 0
	for _, f := range files {
		totalRows += len(f.Rows)
	}
	logger.Info("csv loaded", slog.Int("files", len(files)), slog.Int("rows", totalRows))

	result, err := im.Run(ctx, season, files, maddenimport.Options{Strict: strict})
	if err != nil {
		return err
	}
	logger.Info("import completed",
		slog.Int("updated", result.Updated),
		slog.Int("skipped", result.Skipped),
	)
	return nil
}
