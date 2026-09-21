package domain

// 反則ルールの攻守ステータス（foul_rules.status_type）。
// Laravel の config/const.php の FoulRules と同じ値。
const (
	FoulStatusCommon  int32 = 0
	FoulStatusOffence int32 = 1
	FoulStatusDefence int32 = 2
	FoulStatusKick    int32 = 3
	FoulStatusAll     int32 = 99
)
