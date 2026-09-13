// Package pfr は pro-football-reference.com からロスターを取得・解析する。
package pfr

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// rosterTableSelector はロスター表。ページの JS 初期化が終わると .is_setup が付く。
const rosterTableSelector = ".is_setup #roster"

// ParseRoster はロスターページの HTML からロスター行を抽出する。
// 抽出規則は既存の PHP 実装（ScrapeRosters）と同じ。
func ParseRoster(html string) ([]domain.RawRosterRow, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	table := doc.Find(rosterTableSelector)
	if table.Length() == 0 {
		return nil, fmt.Errorf("roster table not found (selector %q)", rosterTableSelector)
	}

	rows := make([]domain.RawRosterRow, 0, 64)
	table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		// tbody 内に繰り返し挿入される見出し行（th のみ）は選手データを持たないので除外
		if tr.HasClass("thead") || tr.Find(`td[data-stat="player"]`).Length() == 0 {
			return
		}
		rows = append(rows, parseRow(tr))
	})
	return rows, nil
}

func parseRow(tr *goquery.Selection) domain.RawRosterRow {
	var row domain.RawRosterRow

	// 背番号: <th data-stat="uniform_number">
	row.Number = text(tr.Find(`th[data-stat="uniform_number"]`))

	// 選手名: csk 属性 "Last,First"
	player := tr.Find(`td[data-stat="player"]`)
	if csk, ok := player.Attr("csk"); ok {
		row.LastName, row.FirstName = splitName(csk)
	} else {
		// csk が無い場合は表示名 "First Last" から補う
		row.FirstName, row.LastName = splitDisplayName(text(player))
	}

	// ポジション
	row.Position = text(tr.Find(`td[data-stat="pos"]`))
	if row.Position == "" {
		row.Position = "No Data"
	}

	// 生年月日: csk 属性 "YYYY-MM-DD" → "YYYYMMDD"
	if csk, ok := tr.Find(`td[data-stat="birth_date_mod"]`).Attr("csk"); ok {
		row.Birthday = strings.ReplaceAll(csk, "-", "")
	}

	row.WeightLb = text(tr.Find(`td[data-stat="weight"]`))
	row.Height = text(tr.Find(`td[data-stat="height"]`))

	// 大学: <a> があればそのテキスト、無ければ空
	row.College = text(tr.Find(`td[data-stat="college_id"] a`).First())

	// 経験年数: ルーキーは "Rook" と表示されるので "0" にする
	row.Experience = text(tr.Find(`td[data-stat="experience"]`))
	if row.Experience == "Rook" {
		row.Experience = "0"
	}

	// ドラフト情報: "Team / Round / Rank / Year" の形式。年は <a> のテキスト先頭 4 文字
	draft := tr.Find(`td[data-stat="draft_info"]`)
	if dt := text(draft); dt != "" {
		parts := strings.Split(dt, " / ")
		row.DraftTeam = at(parts, 0)
		row.DraftRound = at(parts, 1)
		row.DraftRank = at(parts, 2)
		if y := text(draft.Find("a").First()); len(y) >= 4 {
			row.DraftYear = y[:4]
		}
	}

	return row
}

// splitName は csk 属性 "Last,First" を姓・名に分ける。
func splitName(csk string) (last, first string) {
	parts := strings.SplitN(csk, ",", 2)
	last = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		first = strings.TrimSpace(parts[1])
	}
	return last, first
}

// splitDisplayName は表示名 "First Last" を最後の空白で名・姓に分ける。
func splitDisplayName(name string) (first, last string) {
	name = strings.TrimSpace(name)
	i := strings.LastIndex(name, " ")
	if i < 0 {
		return "", name
	}
	return name[:i], name[i+1:]
}

func text(s *goquery.Selection) string {
	return strings.TrimSpace(s.Text())
}

func at(parts []string, i int) string {
	if i < len(parts) {
		return strings.TrimSpace(parts[i])
	}
	return ""
}
