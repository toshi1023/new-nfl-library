package domain

import "context"

// TeamDetail は検索UI（GET /api/v1/search/teams）が返す読み取り用のチーム1件。
// スクレイピング用の Team（略称のみ）とは別物。
type TeamDetail struct {
	ID            int32
	City          string
	Name          string
	Conference    int16
	Area          int16
	ImageFile     *string
	BackImageFile *string
}

// SearchRepository は検索UIの選択肢（シーズン一覧・チーム一覧）を返す。
type SearchRepository interface {
	// ListSeasons はロスターが存在するシーズンの一覧を降順で返す。
	ListSeasons(ctx context.Context) ([]int32, error)
	// ListTeams は全チームを id 昇順で返す。
	ListTeams(ctx context.Context) ([]TeamDetail, error)
}
