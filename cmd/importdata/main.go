// Command importdata は db/data にある CSV を一括で読み込み、players / rosters / starters テーブルへ
// 反映する。db/data/roster_starters（cmd/scraperosters・cmd/scrapestarters の出力形式）と
// db/data/madden（Madden NFL の評価値CSV。年ごとのサブディレクトリ構成）の両方を対象にできる。
//
// 使い方:
//
//	go run ./cmd/importdata                        # rosters・starters・madden とも全シーズン
//	go run ./cmd/importdata -type rosters          # rosters だけ
//	go run ./cmd/importdata -type starters -season 2024
//	go run ./cmd/importdata -type madden -season 2024
//
// -type all（既定）のときは rosters → starters → madden の順に全シーズン投入する。starters・madden は
// 該当する roster 行が無い行をスキップするため、rosters を先に投入しておく必要がある。
//
// db/data/madden は年ごとのサブディレクトリ（db/data/madden/<season>/）に分かれており、
// サブディレクトリ名がそのままシーズンになる。各サブディレクトリの中身はチームごとに分かれた
// 複数ファイルでも、全チームがまとまった単一ファイルでもよい（cmd/importmadden と同じ扱い）。
//
// 事前に `make seed` で teams / positions が投入済みであること。
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
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"

	"github.com/toshi1023/new-nfl-library/internal/infra/csvwriter"
	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/maddenimport"
	"github.com/toshi1023/new-nfl-library/internal/rosterimport"
	"github.com/toshi1023/new-nfl-library/internal/starterimport"
)

type config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// dataType は -type フラグで選べる対象。
type dataType string

const (
	typeRosters  dataType = "rosters"
	typeStarters dataType = "starters"
	typeMadden   dataType = "madden"
	typeAll      dataType = "all"
)

func main() {
	if err := run(); err != nil {
		slog.Error("import data failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	dir := flag.String("dir", "db/data/roster_starters", "rosters/starters CSVのあるディレクトリ")
	maddenDir := flag.String("madden-dir", "db/data/madden", "Madden評価値CSVのあるディレクトリ（年ごとのサブディレクトリ構成）")
	typeFlag := flag.String("type", string(typeAll), "対象（rosters / starters / madden / all）")
	season := flag.Int("season", 0, "このシーズンだけに絞る（未指定なら全シーズン）")
	strict := flag.Bool("strict", false, "positions マスタに無いポジションが出た時点で処理全体を中断する（既定はスキップして続行）")
	flag.Parse()

	dt := dataType(*typeFlag)
	if dt != typeRosters && dt != typeStarters && dt != typeMadden && dt != typeAll {
		return fmt.Errorf("-type must be one of rosters/starters/madden/all, got %q", *typeFlag)
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx = logging.WithLogger(ctx, logger)

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if dt == typeRosters || dt == typeAll {
		if err := importRosters(ctx, pool, *dir, *season, *strict, logger); err != nil {
			return err
		}
	}
	if dt == typeStarters || dt == typeAll {
		if err := importStarters(ctx, pool, *dir, *season, *strict, logger); err != nil {
			return err
		}
	}
	if dt == typeMadden || dt == typeAll {
		if err := importMadden(ctx, pool, *maddenDir, *season, *strict, logger); err != nil {
			return err
		}
	}
	logger.Info("import data completed")
	return nil
}

// rostersFilePattern / startersFilePattern は db/data のファイル名規約（cmd/scraperosters・
// cmd/scrapestarters の出力ファイル名）に対応する。先頭の数字をシーズンとして取り出す。
var (
	rostersFilePattern  = regexp.MustCompile(`^(\d+)_rosters\.csv$`)
	startersFilePattern = regexp.MustCompile(`^(\d+)_starters\.csv$`)
)

// seasonFile は season が判明した対象ファイル 1 件。
type seasonFile struct {
	Season int
	Path   string
}

// findSeasonFiles は dir 直下から pattern に一致するファイルを探し、season でソートして返す。
// seasonFilter が 0 より大きい場合はそのシーズンだけに絞る。
func findSeasonFiles(dir string, pattern *regexp.Regexp, seasonFilter int) ([]seasonFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var files []seasonFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		season, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if seasonFilter != 0 && season != seasonFilter {
			continue
		}
		files = append(files, seasonFile{Season: season, Path: filepath.Join(dir, e.Name())})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Season < files[j].Season })
	return files, nil
}

func importRosters(ctx context.Context, pool *pgxpool.Pool, dir string, seasonFilter int, strict bool, logger *slog.Logger) error {
	files, err := findSeasonFiles(dir, rostersFilePattern, seasonFilter)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		logger.Warn("no roster files found", slog.String("dir", dir))
		return nil
	}

	im := rosterimport.New(pool)
	var totalCreated, totalReused, totalUpserted, totalSkipped int
	for _, f := range files {
		entries, err := csvwriter.ReadRosterEntriesFile(f.Path)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		res, err := im.Run(ctx, entries, rosterimport.Options{Strict: strict})
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		logger.Info("rosters imported",
			slog.Int("season", f.Season), slog.Int("rows", len(entries)),
			slog.Int("players_created", res.PlayersCreated), slog.Int("players_reused", res.PlayersReused),
			slog.Int("rosters_upserted", res.RostersUpserted), slog.Int("skipped", res.Skipped),
		)
		totalCreated += res.PlayersCreated
		totalReused += res.PlayersReused
		totalUpserted += res.RostersUpserted
		totalSkipped += res.Skipped
	}
	logger.Info("rosters import total",
		slog.Int("seasons", len(files)), slog.Int("players_created", totalCreated), slog.Int("players_reused", totalReused),
		slog.Int("rosters_upserted", totalUpserted), slog.Int("skipped", totalSkipped),
	)
	return nil
}

func importStarters(ctx context.Context, pool *pgxpool.Pool, dir string, seasonFilter int, strict bool, logger *slog.Logger) error {
	files, err := findSeasonFiles(dir, startersFilePattern, seasonFilter)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		logger.Warn("no starter files found", slog.String("dir", dir))
		return nil
	}

	im := starterimport.New(pool)
	var totalUpserted, totalSkipped int
	for _, f := range files {
		entries, err := csvwriter.ReadStarterEntriesFile(f.Path)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		res, err := im.Run(ctx, entries, starterimport.Options{Strict: strict})
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		logger.Info("starters imported",
			slog.Int("season", f.Season), slog.Int("rows", len(entries)),
			slog.Int("upserted", res.Upserted), slog.Int("skipped", res.Skipped),
		)
		totalUpserted += res.Upserted
		totalSkipped += res.Skipped
	}
	logger.Info("starters import total",
		slog.Int("seasons", len(files)), slog.Int("upserted", totalUpserted), slog.Int("skipped", totalSkipped),
	)
	return nil
}

func importMadden(ctx context.Context, pool *pgxpool.Pool, dir string, seasonFilter int, strict bool, logger *slog.Logger) error {
	_, err := maddenimport.New(pool).RunSeasons(ctx, dir, seasonFilter, maddenimport.Options{Strict: strict}, logger)
	return err
}
