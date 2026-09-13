// Package usecase はビジネスロジックを実装する。domain のインターフェースにのみ依存する。
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// MinSeason はスクレイプ対象にできる最古のシーズン。
const MinSeason = 2012

// 単位換算値。既存 DB との整合のため PHP 実装（config/const.php の Calc）と同じ値を使う。
const (
	kgPerLb   = 0.45
	cmPerFoot = 30.48
	cmPerInch = 2.54
)

// ScrapeRosters はロスターをスクレイプして一括出力するユースケース。
type ScrapeRosters struct {
	fetcher domain.RosterFetcher
	writer  domain.RosterWriter
	delay   time.Duration
}

// NewScrapeRosters は ScrapeRosters を生成する。delay はチーム間の待機時間（アクセス集中による遮断を避ける）。
func NewScrapeRosters(fetcher domain.RosterFetcher, writer domain.RosterWriter, delay time.Duration) *ScrapeRosters {
	return &ScrapeRosters{fetcher: fetcher, writer: writer, delay: delay}
}

// Run は teams を順に取得してメモリ上に蓄積し、最後に writer へ一括で渡す。
// 途中で失敗しても、それまでに取得した分は書き出したうえでエラーを返す。
// 戻り値は書き出した行数。
func (u *ScrapeRosters) Run(ctx context.Context, season int, teams []domain.Team) (int, error) {
	if season < MinSeason {
		return 0, fmt.Errorf("season must be %d or later, got %d: %w", MinSeason, season, domain.ErrValidation)
	}
	if len(teams) == 0 {
		return 0, fmt.Errorf("no teams specified: %w", domain.ErrValidation)
	}
	logger := logging.FromContext(ctx)

	entries := make([]domain.RosterEntry, 0, len(teams)*64)
	var runErr error
	for i, team := range teams {
		if i > 0 && u.delay > 0 {
			select {
			case <-ctx.Done():
				runErr = ctx.Err()
			case <-time.After(u.delay):
			}
			if runErr != nil {
				break
			}
		}

		rows, err := u.fetcher.FetchRoster(ctx, season, team)
		if err != nil {
			runErr = fmt.Errorf("team %s: %w", team.Abbr, err)
			break
		}
		for _, row := range rows {
			entries = append(entries, toEntry(season, team, row))
		}
		logger.Info("roster fetched",
			slog.String("team", team.Abbr),
			slog.Int("season", season),
			slog.Int("rows", len(rows)),
			slog.Int("progress", i+1),
			slog.Int("total", len(teams)),
		)
	}

	if len(entries) == 0 {
		return 0, runErr
	}
	if err := u.writer.Write(ctx, entries); err != nil {
		return 0, errors.Join(runErr, fmt.Errorf("write rosters: %w", err))
	}
	return len(entries), runErr
}

// toEntry は生の行を CSV 出力用に変換する。単位換算はここでのみ行う。
func toEntry(season int, team domain.Team, row domain.RawRosterRow) domain.RosterEntry {
	return domain.RosterEntry{
		Season:       season,
		TeamID:       team.ID,
		Team:         team.Abbr,
		Number:       row.Number,
		Position:     row.Position,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Birthday:     row.Birthday,
		WeightKg:     convertWeight(row.WeightLb),
		HeightCm:     convertHeight(row.Height),
		College:      row.College,
		Experience:   row.Experience,
		DraftedTeam:  row.DraftTeam,
		DraftedRound: row.DraftRound,
		DraftedRank:  row.DraftRank,
		DraftedYear:  row.DraftYear,
	}
}

// convertWeight はポンド表記を kg に換算する（小数第 1 位で四捨五入）。空や数値でなければ nil。
func convertWeight(lb string) *float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(lb), 64)
	if err != nil || v == 0 {
		return nil
	}
	return round1(v * kgPerLb)
}

// convertHeight は "6-2"（フィート-インチ）表記を cm に換算する（小数第 1 位で四捨五入）。形式が違えば nil。
func convertHeight(h string) *float64 {
	parts := strings.SplitN(strings.TrimSpace(h), "-", 2)
	if len(parts) != 2 {
		return nil
	}
	feet, err1 := strconv.ParseFloat(parts[0], 64)
	inch, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	return round1(feet*cmPerFoot + inch*cmPerInch)
}

func round1(v float64) *float64 {
	r := math.Round(v*10) / 10
	return &r
}
