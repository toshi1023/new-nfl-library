package domain

import "context"

// RawStarterRow はスターター表 1 行をそのまま文字列で保持したもの。単位変換前。
type RawStarterRow struct {
	Position  string // ポジション（例: QB, LT, RDE）
	FirstName string
	LastName  string
}

// StarterEntry は CSV に出力する 1 行。starters テーブルの情報を players / positions と合わせて 1 行にまとめたもの。
type StarterEntry struct {
	Season    int
	TeamID    int
	Team      string
	Position  string
	FirstName string
	LastName  string
	// Odflg は攻守フラグ（true: Offence, false: Defence）。domain.IsOffenceCategory で判定する。
	// ポジションがマスタに見つからない場合は false（Laravel の ScrapeStarters が
	// position_model が nil のとき odflg を false とするのと同じ挙動）。
	Odflg bool
}

// StarterFetcher は指定シーズン・チームのスターター表を取得する。
type StarterFetcher interface {
	FetchStarters(ctx context.Context, season int, team Team) ([]RawStarterRow, error)
}

// StarterWriter は取得したスターターを一括で出力する。
type StarterWriter interface {
	Write(ctx context.Context, entries []StarterEntry) error
}
