// Package csvwriter はロスターを CSV ファイルに出力する。
package csvwriter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// Header は CSV の 1 行目。
var header = []string{
	"season", "team_id", "team", "number", "position", "firstname", "lastname", "birthday",
	"weight_kg", "height_cm", "college", "experience",
	"drafted_team", "drafted_round", "drafted_rank", "drafted_year",
}

// Writer はロスターを指定パスの CSV に書き出す。domain.RosterWriter を実装する。
type Writer struct {
	path string
}

// NewWriter は出力先パスを受け取って Writer を生成する。
func NewWriter(path string) *Writer {
	return &Writer{path: path}
}

// Write は entries を CSV ファイルとして一括で書き出す。出力先ディレクトリが無ければ作成し、既存ファイルは上書きする。
func (w *Writer) Write(_ context.Context, entries []domain.RosterEntry) (err error) {
	if dir := filepath.Dir(w.path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	f, err := os.Create(w.path)
	if err != nil {
		return fmt.Errorf("create %s: %w", w.path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", w.path, cerr)
		}
	}()

	if err := Encode(f, entries); err != nil {
		return fmt.Errorf("write %s: %w", w.path, err)
	}
	return nil
}

// Encode は entries をヘッダ付き CSV として out に書き込む。
func Encode(out io.Writer, entries []domain.RosterEntry) error {
	cw := csv.NewWriter(out)
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for i, e := range entries {
		if err := cw.Write(record(e)); err != nil {
			return fmt.Errorf("write row %d: %w", i, err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

func record(e domain.RosterEntry) []string {
	return []string{
		strconv.Itoa(e.Season),
		strconv.Itoa(e.TeamID),
		e.Team,
		e.Number,
		e.Position,
		e.FirstName,
		e.LastName,
		e.Birthday,
		formatFloat(e.WeightKg),
		formatFloat(e.HeightCm),
		e.College,
		e.Experience,
		e.DraftedTeam,
		e.DraftedRound,
		e.DraftedRank,
		e.DraftedYear,
	}
}

func formatFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 1, 64)
}
