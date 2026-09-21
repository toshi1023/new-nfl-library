package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// ScrapeStarters はスターターをスクレイプして一括出力するユースケース。
type ScrapeStarters struct {
	fetcher domain.StarterFetcher
	writer  domain.StarterWriter
	delay   time.Duration
}

// NewScrapeStarters は ScrapeStarters を生成する。delay はチーム間の待機時間。
func NewScrapeStarters(fetcher domain.StarterFetcher, writer domain.StarterWriter, delay time.Duration) *ScrapeStarters {
	return &ScrapeStarters{fetcher: fetcher, writer: writer, delay: delay}
}

// Run は teams を順に取得してメモリ上に蓄積し、最後に writer へ一括で渡す。
// 途中で失敗しても、それまでに取得した分は書き出したうえでエラーを返す。
// 戻り値は書き出した行数。
func (u *ScrapeStarters) Run(ctx context.Context, season int, teams []domain.Team) (int, error) {
	if season < MinSeason {
		return 0, fmt.Errorf("season must be %d or later, got %d: %w", MinSeason, season, domain.ErrValidation)
	}
	if len(teams) == 0 {
		return 0, fmt.Errorf("no teams specified: %w", domain.ErrValidation)
	}
	logger := logging.FromContext(ctx)

	entries := make([]domain.StarterEntry, 0, len(teams)*22)
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

		rows, err := u.fetcher.FetchStarters(ctx, season, team)
		if err != nil {
			runErr = fmt.Errorf("team %s: %w", team.Abbr, err)
			break
		}
		for _, row := range rows {
			entry := toStarterEntry(season, team, row)
			entries = append(entries, entry)
			logger.Debug("starter", slog.String("team", team.Abbr), slog.String("position", entry.Position),
				slog.String("firstname", entry.FirstName), slog.String("lastname", entry.LastName))
		}
		logger.Info("starters fetched",
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
		return 0, errors.Join(runErr, fmt.Errorf("write starters: %w", err))
	}
	return len(entries), runErr
}

// toStarterEntry は生の行を CSV 出力用に変換する。odflg はポジション名からカテゴリを引いて判定する
// （domain.IsOffenceCategory）。ポジションがマスタに無ければ false（Laravel の実装が position_model が
// nil のとき odflg を false とするのと同じ挙動）。
func toStarterEntry(season int, team domain.Team, row domain.RawStarterRow) domain.StarterEntry {
	var odflg bool
	if category, ok := domain.PositionCategory(row.Position); ok {
		odflg = domain.IsOffenceCategory(category)
	}
	return domain.StarterEntry{
		Season:    season,
		TeamID:    team.ID,
		Team:      team.Abbr,
		Position:  row.Position,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Odflg:     odflg,
	}
}
