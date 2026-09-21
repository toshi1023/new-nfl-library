package pfr

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// startersTableSelector はスターター表。ロスターページ内に #roster とは別に存在する。
const startersTableSelector = "#starters"

// ParseStarters はロスターページの HTML からスターター表の行を抽出する。
// 抽出規則は既存の PHP 実装（ScrapeStarters）と同じ。攻撃/守備の見出し行（class="full_table" が付かない行）は除外する。
func ParseStarters(html string) ([]domain.RawStarterRow, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	table := doc.Find(startersTableSelector)
	if table.Length() == 0 {
		return nil, fmt.Errorf("starters table not found (selector %q)", startersTableSelector)
	}

	rows := make([]domain.RawStarterRow, 0, 22)
	table.Find("tbody tr.full_table").Each(func(_ int, tr *goquery.Selection) {
		rows = append(rows, parseStarterRow(tr))
	})
	return rows, nil
}

func parseStarterRow(tr *goquery.Selection) domain.RawStarterRow {
	var row domain.RawStarterRow

	// ポジション: <th data-stat="pos"> のテキストそのもの
	row.Position = text(tr.Find(`th[data-stat="pos"]`))

	// 選手名: csk 属性 "Last,First"（無ければ表示名の最後の空白で分割）
	player := tr.Find(`td[data-stat="player"]`)
	if csk, ok := player.Attr("csk"); ok {
		row.LastName, row.FirstName = splitName(csk)
	} else {
		row.FirstName, row.LastName = splitDisplayName(text(player))
	}

	return row
}
