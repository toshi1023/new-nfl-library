package csvwriter

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// ReadStarterEntriesFile は path の CSV ファイルを読み込み、[]domain.StarterEntry にする。
// EncodeStarters で書き出した形式（ヘッダ行 + startersHeader で定義した列順）だけを受け付ける。
func ReadStarterEntriesFile(path string) ([]domain.StarterEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	entries, err := DecodeStarterEntries(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return entries, nil
}

// DecodeStarterEntries は EncodeStarters が書き出した CSV を読み、[]domain.StarterEntry に変換する。
func DecodeStarterEntries(in io.Reader) ([]domain.StarterEntry, error) {
	cr := csv.NewReader(in)

	head, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if len(head) != len(startersHeader) {
		return nil, fmt.Errorf("unexpected header: got %d columns, want %d", len(head), len(startersHeader))
	}
	for i, want := range startersHeader {
		if head[i] != want {
			return nil, fmt.Errorf("unexpected header column %d: got %q, want %q", i, head[i], want)
		}
	}

	var entries []domain.StarterEntry
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", len(entries)+1, err)
		}
		e, err := parseStarterRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", len(entries)+1, err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func parseStarterRecord(rec []string) (domain.StarterEntry, error) {
	season, err := strconv.Atoi(rec[0])
	if err != nil {
		return domain.StarterEntry{}, fmt.Errorf("season %q: %w", rec[0], err)
	}
	teamID, err := strconv.Atoi(rec[1])
	if err != nil {
		return domain.StarterEntry{}, fmt.Errorf("team_id %q: %w", rec[1], err)
	}
	odflg, err := strconv.ParseBool(rec[6])
	if err != nil {
		return domain.StarterEntry{}, fmt.Errorf("odflg %q: %w", rec[6], err)
	}

	return domain.StarterEntry{
		Season:    season,
		TeamID:    teamID,
		Team:      rec[2],
		Position:  rec[3],
		FirstName: rec[4],
		LastName:  rec[5],
		Odflg:     odflg,
	}, nil
}
