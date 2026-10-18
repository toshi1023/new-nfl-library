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

// RosterRepository は season・team_id・選手名でロスターを検索する。
type RosterRepository interface {
	// ListBySeasonAndTeam は season・team_id に一致する roster を返す。names が空でなければ、
	// いずれか1語でも firstname・lastname に部分一致（大文字小文字区別なし）する roster に絞り込む
	// （複数指定時は OR 条件。例: ["Josh", "Diggs"] なら Josh Allen と Stefon Diggs の両方が対象）。
	// sortBy は "rating"/"number"/"experience"/"name" のいずれか、sortOrder は "asc"/"desc" のいずれか。
	// number・experience・rating が NULL の roster は並び順（昇順・降順いずれでも）に関わらず常に最後になる。
	ListBySeasonAndTeam(ctx context.Context, season, teamID int, names []string, sortBy, sortOrder string) ([]RosterDetail, error)
}
