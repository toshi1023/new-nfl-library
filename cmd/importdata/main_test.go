package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeEmpty(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("season\n"), 0o644))
}

func TestFindSeasonFiles(t *testing.T) {
	dir := t.TempDir()
	writeEmpty(t, dir, "2012_rosters.csv")
	writeEmpty(t, dir, "2024_rosters.csv")
	writeEmpty(t, dir, "2013_rosters.csv")
	writeEmpty(t, dir, "2024_starters.csv")
	writeEmpty(t, dir, "notes.txt")
	require.NoError(t, os.Mkdir(filepath.Join(dir, "2024_rosters.csv_dir"), 0o755))

	t.Run("rosters, all seasons, sorted ascending", func(t *testing.T) {
		files, err := findSeasonFiles(dir, rostersFilePattern, 0)
		require.NoError(t, err)
		require.Len(t, files, 3)
		assert.Equal(t, []int{2012, 2013, 2024}, []int{files[0].Season, files[1].Season, files[2].Season})
	})

	t.Run("starters only match starters files", func(t *testing.T) {
		files, err := findSeasonFiles(dir, startersFilePattern, 0)
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, 2024, files[0].Season)
	})

	t.Run("season filter", func(t *testing.T) {
		files, err := findSeasonFiles(dir, rostersFilePattern, 2013)
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, 2013, files[0].Season)
		assert.Equal(t, filepath.Join(dir, "2013_rosters.csv"), files[0].Path)
	})

	t.Run("no match for unrelated files", func(t *testing.T) {
		files, err := findSeasonFiles(dir, rostersFilePattern, 9999)
		require.NoError(t, err)
		assert.Empty(t, files)
	})
}

func TestFindSeasonFiles_MissingDir(t *testing.T) {
	_, err := findSeasonFiles(filepath.Join(t.TempDir(), "does-not-exist"), rostersFilePattern, 0)
	require.Error(t, err)
}
