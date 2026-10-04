// Package maddenimport は Madden NFL の評価値CSVを読み込み、rosters テーブルを更新する。
// Laravel の CsvImport コマンドに相当する。
package maddenimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// Row は Madden CSV 1 行分の情報。列の意味が分かった項目だけを個別フィールドに取り出し、
// それ以外の列はすべて Extra（ヘッダー名 → 値）にまとめる。
type Row struct {
	// Team は CSV に Team 列がある場合のその値（"Bills" または "Buffalo Bills" 形式）。
	// 列自体が無い CSV（2014・2015年分）では空文字になる。
	Team string
	// Position は Position ID 列があればその値（短縮コード）、無ければ Position 列の値。
	Position  string
	FirstName string
	LastName  string
	// Number は背番号列があればその値。列が無い、またはセルが空なら空文字（既存値を保持する合図）。
	Number string
	Rating string
	// Extra は上記に使わなかった残り全列。ヘッダー名をキーにした値。
	Extra map[string]string
}

// ParseCSV は Madden の評価値CSVを読み、行ごとの Row を返す。
// 列名の表記ゆれ（年度・Madden のバージョンによって異なる）は正規化して吸収する。
// hasTeamColumn は CSV に Team 列があったかどうか（無ければ呼び出し側でファイル名等から補う必要がある）。
func ParseCSV(r io.Reader) (rows []Row, hasTeamColumn bool, err error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // 実データは揃っているが、念のため列数のブレを許容する

	head, err := cr.Read()
	if err != nil {
		return nil, false, fmt.Errorf("read header: %w", err)
	}
	cols := detectColumns(head)

	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("read row %d: %w", len(rows)+1, err)
		}
		rows = append(rows, buildRow(head, rec, cols))
	}
	return rows, cols.team >= 0, nil
}

// columns はヘッダーから検出した列インデックス。無ければ -1。
type columns struct {
	team, position, positionID, firstName, lastName, name, jersey, rating int
}

// detectColumns はヘッダーを正規化してキーワード判定し、既知の列を検出する。
// PHP 実装（固定の別名リスト）と異なり正規化＋キーワード方式にしているのは、
// 実データで別名リストに無い表記（例: "jerseyNum"）が既に複数見つかっているため。
//
// 各フィールドは最初に一致した列だけを採用する（先勝ち）。2024年分の CSV には主レーティング列
// "Overall" の他に "stats/overall/value" のような入れ子形式の列があり、後者を rating 列として
// 誤検出しないよう、"/" を含む列は jersey・rating の部分一致判定から除外している。
func detectColumns(head []string) columns {
	c := columns{team: -1, position: -1, positionID: -1, firstName: -1, lastName: -1, name: -1, jersey: -1, rating: -1}
	for i, h := range head {
		n := normalizeHeader(h)
		nested := strings.Contains(h, "/") // "stats/overall/value" のような入れ子形式の列
		switch {
		case n == "team" && c.team < 0:
			c.team = i
		case n == "positionid" && c.positionID < 0:
			c.positionID = i
		case n == "position" && c.position < 0:
			c.position = i
		case (n == "firstname" || n == "first") && c.firstName < 0:
			c.firstName = i
		case (n == "lastname" || n == "last") && c.lastName < 0:
			c.lastName = i
		case (n == "name" || n == "fullname") && c.name < 0:
			c.name = i
		case strings.Contains(n, "jersey") && !nested && c.jersey < 0:
			c.jersey = i
		case (n == "ovr" || strings.Contains(n, "overall")) && !nested && c.rating < 0:
			c.rating = i
		}
	}
	return c
}

// normalizeHeader は列名を英数字だけの小文字に正規化する。
// 大文字/小文字・空白・記号の違いを吸収し、UTF-8 BOM（実データのヘッダー先頭に付く）も取り除く。
func normalizeHeader(h string) string {
	h = strings.TrimPrefix(h, "\ufeff")
	var b strings.Builder
	for _, r := range h {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		}
	}
	return b.String()
}

func buildRow(head, rec []string, c columns) Row {
	var row Row
	row.Extra = make(map[string]string)
	used := make(map[int]bool)

	get := func(i int) string {
		if i < 0 || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}

	if c.team >= 0 {
		row.Team = get(c.team)
		used[c.team] = true
	}

	// 氏名: 結合列（Name/Full Name）があれば優先し、分割列があっても使わない
	// （Laravel の CsvImport も同じ優先順位。2022年分は両方の形式が同じファイルに存在する）。
	if c.name >= 0 {
		row.FirstName, row.LastName = splitDisplayName(get(c.name))
		used[c.name] = true
	} else {
		row.FirstName = get(c.firstName)
		// 別々の列でも Madden 側は姓に世代呼称を含めることがある（例: "Bates III"）ため取り除く
		row.LastName = stripNameSuffix(get(c.lastName))
		used[c.firstName] = true
		used[c.lastName] = true
	}

	// ポジション: Position ID があれば優先する（2025年分は Position がフルワード表記になったため）。
	// Position 列自体は消費した扱いにせず、存在すれば Extra に残す。
	if c.positionID >= 0 {
		row.Position = get(c.positionID)
		used[c.positionID] = true
	} else {
		row.Position = get(c.position)
		used[c.position] = true
	}

	if c.jersey >= 0 {
		row.Number = get(c.jersey)
		used[c.jersey] = true
	}
	if c.rating >= 0 {
		row.Rating = get(c.rating)
		used[c.rating] = true
	}

	for i, h := range head {
		if used[i] || i >= len(rec) {
			continue
		}
		key := strings.TrimSpace(strings.TrimPrefix(h, "\ufeff"))
		if key == "" {
			continue
		}
		row.Extra[key] = strings.TrimSpace(rec[i])
	}

	return row
}

// nameSuffixes は氏名末尾に付く世代呼称。スクレイプ元（PFR）の players.lastname にはこれらが
// 含まれない（例: "Antoine Winfield Jr." は firstname="Antoine", lastname="Winfield"）ため、
// Madden CSV 側の表記と突き合わせるために取り除く対象として保持する。
var nameSuffixes = map[string]bool{
	"jr": true, "jr.": true,
	"sr": true, "sr.": true,
	"ii": true, "iii": true, "iv": true, "v": true,
}

// stripNameSuffix は氏名末尾の世代呼称（Jr, Sr, II, III, IV, V。ピリオド有無どちらも可）を取り除く。
func stripNameSuffix(name string) string {
	name = strings.TrimSpace(name)
	i := strings.LastIndex(name, " ")
	if i < 0 {
		return name
	}
	if last := name[i+1:]; nameSuffixes[strings.ToLower(last)] {
		return strings.TrimSpace(name[:i])
	}
	return name
}

// splitDisplayName は "First Last" 形式の氏名を最後の空白で名・姓に分ける。
// PHP 実装は先頭の空白で分割するため Jr. などのサフィックス付き氏名で誤動作するが、
// ここでは internal/infra/pfr のフォールバックと同じ「最後の空白で区切る」方式にしている。
// 分割の前に世代呼称を取り除くので、"Antoine Winfield Jr." は first="Antoine", last="Winfield" になる。
func splitDisplayName(name string) (first, last string) {
	name = stripNameSuffix(strings.TrimSpace(name))
	i := strings.LastIndex(name, " ")
	if i < 0 {
		return "", name
	}
	return name[:i], name[i+1:]
}

// TeamHintFromFilename はファイル名の先頭トークン（"_" 区切り）をチーム略称とみなして解決する。
// Team 列を持たない CSV（2014・2015年分）で、ファイル名からチームを判定するために使う
// （例: "atl_madden_nfl_15.csv" → atl → Atlanta Falcons）。
func TeamHintFromFilename(path string) (domain.Team, bool) {
	base := filepath.Base(path)
	abbr, _, ok := strings.Cut(base, "_")
	if !ok {
		return domain.Team{}, false
	}
	t, err := domain.FindTeam(strings.ToLower(abbr))
	if err != nil {
		return domain.Team{}, false
	}
	return t, true
}
