package csvwriter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// startersHeader は starters CSV の 1 行目。
var startersHeader = []string{"season", "team_id", "team", "position", "firstname", "lastname", "odflg"}

// StartersWriter はスターターを指定パスの CSV に書き出す。domain.StarterWriter を実装する。
type StartersWriter struct {
	path string
}

// NewStartersWriter は出力先パスを受け取って StartersWriter を生成する。
func NewStartersWriter(path string) *StartersWriter {
	return &StartersWriter{path: path}
}

// Write は entries を CSV ファイルとして一括で書き出す。出力先ディレクトリが無ければ作成し、既存ファイルは上書きする。
func (w *StartersWriter) Write(_ context.Context, entries []domain.StarterEntry) error {
	return writeFile(w.path, func(out io.Writer) error { return EncodeStarters(out, entries) })
}

// EncodeStarters は entries をヘッダ付き CSV として out に書き込む。
func EncodeStarters(out io.Writer, entries []domain.StarterEntry) error {
	cw := csv.NewWriter(out)
	if err := cw.Write(startersHeader); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for i, e := range entries {
		if err := cw.Write(startersRecord(e)); err != nil {
			return fmt.Errorf("write row %d: %w", i, err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

func startersRecord(e domain.StarterEntry) []string {
	return []string{
		strconv.Itoa(e.Season),
		strconv.Itoa(e.TeamID),
		e.Team,
		e.Position,
		e.FirstName,
		e.LastName,
		strconv.FormatBool(e.Odflg),
	}
}
