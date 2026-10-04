package maddenimport

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// CollectPaths は explicit（-csv 等で明示されたパス）と dir（指定すればそのディレクトリ直下の
// *.csv をすべて対象にする。サブディレクトリは見ない）から対象ファイル一覧を組み立てる。
// dir が空文字なら explicit だけを返す。
func CollectPaths(dir string, explicit []string) ([]string, error) {
	paths := append([]string{}, explicit...)
	if dir == "" {
		return paths, nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", dir, err)
	}
	return append(paths, matches...), nil
}

// LoadFiles は paths の各 CSV を読み込み、Team 列が無いファイルには TeamHintFromFilename による
// チーム推定を付与する。cmd/importmadden・cmd/importdata の両方から使われる。
func LoadFiles(paths []string, logger *slog.Logger) ([]File, error) {
	files := make([]File, 0, len(paths))
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}
		rows, hasTeam, err := ParseCSV(f)
		closeErr := f.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s: %w", path, closeErr)
		}

		file := File{Path: path, Rows: rows}
		if hasTeam {
			files = append(files, file)
			continue
		}
		// Team 列が無い CSV はファイル名からチームを推定する
		team, ok := TeamHintFromFilename(path)
		if !ok {
			return nil, fmt.Errorf("%s has no Team column and the team could not be determined from the filename", path)
		}
		file.FallbackTeamID = int32(team.ID)
		logger.Debug("resolved team from filename", slog.String("path", path), slog.String("team", team.Abbr))
		files = append(files, file)
	}
	return files, nil
}

// seasonDirPattern は <rootDir>/<season>/ のサブディレクトリ名（4桁の年）に対応する
// （例: db/data/madden/2024/）。
var seasonDirPattern = regexp.MustCompile(`^\d{4}$`)

// SeasonDir は season が判明した対象ディレクトリ 1 件。
type SeasonDir struct {
	Season int
	Path   string
}

// FindSeasonDirs は rootDir 直下のサブディレクトリのうち、名前が4桁の年であるものを season で
// ソートして返す。seasonFilter が 0 より大きい場合はそのシーズンだけに絞る。
// db/data/madden のような「年ごとのサブディレクトリに分かれたデータセット」を一括処理する際に使う。
func FindSeasonDirs(rootDir string, seasonFilter int) ([]SeasonDir, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", rootDir, err)
	}

	var dirs []SeasonDir
	for _, e := range entries {
		if !e.IsDir() || !seasonDirPattern.MatchString(e.Name()) {
			continue
		}
		season, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if seasonFilter != 0 && season != seasonFilter {
			continue
		}
		dirs = append(dirs, SeasonDir{Season: season, Path: filepath.Join(rootDir, e.Name())})
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Season < dirs[j].Season })
	return dirs, nil
}

// RunSeasons は rootDir 配下の全シーズン（FindSeasonDirs で見つかったもの。seasonFilter が 0 より
// 大きければそのシーズンだけ）に対して Run を実行し、結果を合計する。シーズンごとに INFO ログを、
// 最後に合計の INFO ログを出す。cmd/importmadden（-season 省略時）・cmd/importdata の両方から使う。
func (im *Importer) RunSeasons(ctx context.Context, rootDir string, seasonFilter int, opts Options, logger *slog.Logger) (Result, error) {
	dirs, err := FindSeasonDirs(rootDir, seasonFilter)
	if err != nil {
		return Result{}, err
	}
	if len(dirs) == 0 {
		logger.Warn("no madden season directories found", slog.String("dir", rootDir))
		return Result{}, nil
	}

	var total Result
	for _, d := range dirs {
		paths, err := CollectPaths(d.Path, nil)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", d.Path, err)
		}
		if len(paths) == 0 {
			logger.Warn("no csv files found", slog.String("dir", d.Path))
			continue
		}
		files, err := LoadFiles(paths, logger)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", d.Path, err)
		}

		res, err := im.Run(ctx, d.Season, files, opts)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", d.Path, err)
		}
		logger.Info("madden imported",
			slog.Int("season", d.Season), slog.Int("files", len(files)),
			slog.Int("updated", res.Updated), slog.Int("skipped", res.Skipped),
		)
		total.Updated += res.Updated
		total.Skipped += res.Skipped
	}
	logger.Info("madden import total",
		slog.Int("seasons", len(dirs)), slog.Int("updated", total.Updated), slog.Int("skipped", total.Skipped),
	)
	return total, nil
}
