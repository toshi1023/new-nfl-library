package domain

import "context"

// RosterDetail は API（/api/v1/rosters）が返す読み取り用のロスター1件。
// player・position を JOIN 済みで保持する。CSV取り込み用の RosterEntry とは別物。
type RosterDetail struct {
	ID         int32
	Season     int32
	TeamID     int32
	PlayerID   int32
	PositionID int32
	Number     *int32
	Rating     *int32
	Experience *int32
	Player     Player
	Position   PositionDetail
}

// RosterProfile は API（/api/v1/rosters/{id}）が返すロスター1件の詳細。
// 一覧用の RosterDetail に所属チーム・スターター情報・Madden 評価値・在籍履歴を加えたもの。
type RosterProfile struct {
	RosterDetail
	Team TeamDetail
	// Starter はスターター情報。スターターでなければ nil。
	Starter *RosterStarter
	// MaddenAttributes は rosters.madden_attributes の生キー（CSVヘッダーそのまま）→値。
	// 日本語名への変換は handler 側で行う（internal/maddenimport.JapaneseLabel）。
	MaddenAttributes map[string]string
	// Career は同じ選手の全シーズンのロスター（season の降順。自身も含む）。
	Career []RosterCareer
}

// RosterStarter はロスターのスターター情報（starters テーブル）。
type RosterStarter struct {
	// Odflg は攻守フラグ（true: Offence, false: Defence）。
	Odflg bool
}

// RosterCareer は選手の在籍履歴1件（あるシーズン・チームでのロスター）。
type RosterCareer struct {
	RosterID     int32
	Season       int32
	Team         TeamDetail
	PositionName string
	Number       *int32
	Rating       *int32
	Experience   *int32
	IsStarter    bool
}

// RosterRepository は season・team_id・選手名でロスターを検索する。
type RosterRepository interface {
	// ListBySeasonAndTeam は season・team_id に一致する roster を返す。names が空でなければ、
	// いずれか1語でも firstname・lastname に部分一致（大文字小文字区別なし）する roster に絞り込む
	// （複数指定時は OR 条件。例: ["Josh", "Diggs"] なら Josh Allen と Stefon Diggs の両方が対象）。
	// sortBy は "rating"/"number"/"experience"/"name" のいずれか、sortOrder は "asc"/"desc" のいずれか。
	// number・experience・rating が NULL の roster は並び順（昇順・降順いずれでも）に関わらず常に最後になる。
	ListBySeasonAndTeam(ctx context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]RosterDetail, error)
	// GetByID は rosters.id に一致する roster を所属チーム・スターター情報・Madden 評価値と合わせて返す。
	// Career は含まない（ListCareerByPlayerID で別途取得する）。存在しなければ ErrNotFound をラップして返す。
	GetByID(ctx context.Context, id int) (RosterProfile, error)
	// ListCareerByPlayerID は player_id に一致する全シーズンの roster を season の降順で返す。
	ListCareerByPlayerID(ctx context.Context, playerID int) ([]RosterCareer, error)
}
