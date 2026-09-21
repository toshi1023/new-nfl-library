package domain

// Team は NFL のチーム。Abbr はスクレイプ先 URL で使う略称、ID は既存 DB の teams.id と対応する。
type Team struct {
	ID   int
	Abbr string
}

// AllTeams は全 32 チームを ID 順に返す。並びは既存の teams テーブルと同じ。
// 32 チームの略称はスクレイプ先の URL に使われるものと同じ。
func AllTeams() []Team {
	return []Team{
		{1, "buf"}, {2, "nwe"}, {3, "mia"}, {4, "nyj"}, // AFC 東
		{5, "cin"}, {6, "pit"}, {7, "cle"}, {8, "rav"}, // AFC 北
		{9, "oti"}, {10, "clt"}, {11, "htx"}, {12, "jax"}, // AFC 南
		{13, "kan"}, {14, "rai"}, {15, "sdg"}, {16, "den"}, // AFC 西
		{17, "dal"}, {18, "phi"}, {19, "was"}, {20, "nyg"}, // NFC 東
		{21, "gnb"}, {22, "min"}, {23, "chi"}, {24, "det"}, // NFC 北
		{25, "tam"}, {26, "nor"}, {27, "atl"}, {28, "car"}, // NFC 南
		{29, "ram"}, {30, "crd"}, {31, "sfo"}, {32, "sea"}, // NFC 西
	}
}

// FindTeam は略称からチームを探す。見つからなければ ErrNotFound をラップして返す。
func FindTeam(abbr string) (Team, error) {
	for _, t := range AllTeams() {
		if t.Abbr == abbr {
			return t, nil
		}
	}
	return Team{}, ErrNotFound
}
