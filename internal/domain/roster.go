package domain

import "context"

// RawRosterRow はスクレイプ先のロスター表 1 行をそのまま文字列で保持したもの。単位変換前。
type RawRosterRow struct {
	Number     string // 背番号
	Position   string // ポジション（空なら "No Data"）
	FirstName  string
	LastName   string
	Birthday   string // YYYYMMDD
	WeightLb   string // ポンド
	Height     string // "6-2" 形式（フィート-インチ）
	College    string
	Experience string // 経験年数。ルーキーは "0"
	DraftTeam  string
	DraftRound string
	DraftRank  string
	DraftYear  string
}

// RosterEntry は CSV に出力する 1 行。players テーブルと rosters テーブルの情報を 1 行にまとめたもの。
type RosterEntry struct {
	Season       int
	TeamID       int
	Team         string
	Number       string
	Position     string
	FirstName    string
	LastName     string
	Birthday     string
	WeightKg     *float64 // 取得できなければ nil
	HeightCm     *float64 // 取得できなければ nil
	College      string
	Experience   string
	DraftedTeam  string
	DraftedRound string
	DraftedRank  string
	DraftedYear  string
}

// RosterFetcher は指定シーズン・チームのロスター表を取得する。
type RosterFetcher interface {
	FetchRoster(ctx context.Context, season int, team Team) ([]RawRosterRow, error)
}

// RosterWriter は取得したロスターを一括で出力する。
type RosterWriter interface {
	Write(ctx context.Context, entries []RosterEntry) error
}
