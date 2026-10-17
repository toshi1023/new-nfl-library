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
	// MaddenAttributes は rosters.madden_attributes の生キー（CSVヘッダーそのまま）→値。
	// 日本語名への変換は handler 側で行う（internal/maddenimport.JapaneseLabel）。
	MaddenAttributes map[string]string
}

// RosterRepository は season・team_id でロスターを検索する。
type RosterRepository interface {
	ListBySeasonAndTeam(ctx context.Context, season, teamID int) ([]RosterDetail, error)
}
