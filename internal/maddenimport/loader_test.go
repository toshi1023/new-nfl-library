package maddenimport

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeCSVFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestCollectPaths(t *testing.T) {
	dir := t.TempDir()
	writeCSVFile(t, filepath.Join(dir, "a.csv"), "x")
	writeCSVFile(t, filepath.Join(dir, "b.csv"), "x")
	writeCSVFile(t, filepath.Join(dir, "notes.txt"), "x")

	t.Run("dir only", func(t *testing.T) {
		paths, err := CollectPaths(dir, nil)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{filepath.Join(dir, "a.csv"), filepath.Join(dir, "b.csv")}, paths)
	})

	t.Run("explicit only", func(t *testing.T) {
		paths, err := CollectPaths("", []string{"x.csv", "y.csv"})
		require.NoError(t, err)
		assert.Equal(t, []string{"x.csv", "y.csv"}, paths)
	})

	t.Run("explicit and dir combined", func(t *testing.T) {
		paths, err := CollectPaths(dir, []string{"z.csv"})
		require.NoError(t, err)
		assert.Len(t, paths, 3)
		assert.Contains(t, paths, "z.csv")
	})
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLoadFiles(t *testing.T) {
	dir := t.TempDir()

	withTeam := filepath.Join(dir, "with_team.csv")
	writeCSVFile(t, withTeam, "Team,Position,First Name,Last Name,Overall\nBills,QB,Josh,Allen,91\n")

	// Team 列が無いファイル。ファイル名の先頭トークン (atl) からチームを推定できる。
	noTeam := filepath.Join(dir, "atl_madden_nfl_15.csv")
	writeCSVFile(t, noTeam, "FIRST,LAST,POSITION,OVERALL RATING\nMatt,Ryan,QB,90\n")

	files, err := LoadFiles([]string{withTeam, noTeam}, discardLogger())
	require.NoError(t, err)
	require.Len(t, files, 2)

	assert.Equal(t, withTeam, files[0].Path)
	assert.Zero(t, files[0].FallbackTeamID, "Team列があるのでフォールバック不要")

	assert.Equal(t, noTeam, files[1].Path)
	assert.NotZero(t, files[1].FallbackTeamID, "Team列が無いのでファイル名から推定する")
}

func TestLoadFiles_NoTeamColumnAndUnresolvableFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown-file.csv")
	writeCSVFile(t, path, "FIRST,LAST,POSITION,OVERALL RATING\nMatt,Ryan,QB,90\n")

	_, err := LoadFiles([]string{path}, discardLogger())
	require.Error(t, err)
}

func TestLoadFiles_MissingFile(t *testing.T) {
	_, err := LoadFiles([]string{filepath.Join(t.TempDir(), "missing.csv")}, discardLogger())
	require.Error(t, err)
}

func mkdirAll(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.Mkdir(filepath.Join(dir, name), 0o755))
}

func TestFindSeasonDirs(t *testing.T) {
	dir := t.TempDir()
	mkdirAll(t, dir, "2012")
	mkdirAll(t, dir, "2024")
	mkdirAll(t, dir, "2013")
	mkdirAll(t, dir, "not-a-year")
	writeCSVFile(t, filepath.Join(dir, "2099.csv"), "x") // ファイルは無視する（ディレクトリのみ対象）

	t.Run("all seasons, sorted ascending", func(t *testing.T) {
		dirs, err := FindSeasonDirs(dir, 0)
		require.NoError(t, err)
		require.Len(t, dirs, 3)
		assert.Equal(t, []int{2012, 2013, 2024}, []int{dirs[0].Season, dirs[1].Season, dirs[2].Season})
	})

	t.Run("season filter", func(t *testing.T) {
		dirs, err := FindSeasonDirs(dir, 2013)
		require.NoError(t, err)
		require.Len(t, dirs, 1)
		assert.Equal(t, filepath.Join(dir, "2013"), dirs[0].Path)
	})

	t.Run("no match for unrelated season", func(t *testing.T) {
		dirs, err := FindSeasonDirs(dir, 9999)
		require.NoError(t, err)
		assert.Empty(t, dirs)
	})
}

func TestFindSeasonDirs_MissingDir(t *testing.T) {
	_, err := FindSeasonDirs(filepath.Join(t.TempDir(), "does-not-exist"), 0)
	require.Error(t, err)
}
