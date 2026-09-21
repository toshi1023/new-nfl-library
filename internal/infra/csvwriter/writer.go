package csvwriter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// header は CSV の 1 行目。
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
func (w *Writer) Write(_ context.Context, entries []domain.RosterEntry) error {
	return writeFile(w.path, func(out io.Writer) error { return Encode(out, entries) })
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
