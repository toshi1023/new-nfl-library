package domain

// Player は選手1人分の情報。API で返す読み取り用エンティティ（CSV取り込み用の RosterEntry とは別物）。
type Player struct {
	ID           int32
	FirstName    string
	LastName     string
	Birthday     string // YYYYMMDD
	HeightCm     *float64
	WeightKg     *float64
	College      *string
	DraftedTeam  *string
	DraftedRound *string
	DraftedRank  *string
	DraftedYear  *int32
	ImageFile    *string
}
