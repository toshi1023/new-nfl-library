package csvwriter

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// ReadRosterEntriesFile は path の CSV ファイルを読み込み、[]domain.RosterEntry にする。
// Encode で書き出した形式（ヘッダ行 + header で定義した列順）だけを受け付ける。
func ReadRosterEntriesFile(path string) ([]domain.RosterEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	entries, err := DecodeRosterEntries(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return entries, nil
}

// DecodeRosterEntries は Encode が書き出した CSV を読み、[]domain.RosterEntry に変換する。
func DecodeRosterEntries(in io.Reader) ([]domain.RosterEntry, error) {
	cr := csv.NewReader(in)

	head, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if len(head) != len(header) {
		return nil, fmt.Errorf("unexpected header: got %d columns, want %d", len(head), len(header))
	}
	for i, want := range header {
		if head[i] != want {
			return nil, fmt.Errorf("unexpected header column %d: got %q, want %q", i, head[i], want)
		}
	}

	var entries []domain.RosterEntry
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", len(entries)+1, err)
		}
		e, err := parseRosterRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", len(entries)+1, err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func parseRosterRecord(rec []string) (domain.RosterEntry, error) {
	season, err := strconv.Atoi(rec[0])
	if err != nil {
		return domain.RosterEntry{}, fmt.Errorf("season %q: %w", rec[0], err)
	}
	teamID, err := strconv.Atoi(rec[1])
	if err != nil {
		return domain.RosterEntry{}, fmt.Errorf("team_id %q: %w", rec[1], err)
	}
	weightKg, err := parseOptionalFloat(rec[8])
	if err != nil {
		return domain.RosterEntry{}, fmt.Errorf("weight_kg %q: %w", rec[8], err)
	}
	heightCm, err := parseOptionalFloat(rec[9])
	if err != nil {
		return domain.RosterEntry{}, fmt.Errorf("height_cm %q: %w", rec[9], err)
	}

	return domain.RosterEntry{
		Season:       season,
		TeamID:       teamID,
		Team:         rec[2],
		Number:       rec[3],
		Position:     rec[4],
		FirstName:    rec[5],
		LastName:     rec[6],
		Birthday:     rec[7],
		WeightKg:     weightKg,
		HeightCm:     heightCm,
		College:      rec[10],
		Experience:   rec[11],
		DraftedTeam:  rec[12],
		DraftedRound: rec[13],
		DraftedRank:  rec[14],
		DraftedYear:  rec[15],
	}, nil
}

func parseOptionalFloat(s string) (*float64, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
