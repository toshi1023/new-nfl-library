package maddenimport

import (
	"regexp"
	"strings"
)

// nestedStatPattern は2024年分の入れ子形式のヘッダー（例: "stats/speed/value"）に対応する。
var nestedStatPattern = regexp.MustCompile(`^stats/([A-Za-z]+)/(value|diff)$`)

// canonicalStatKey は madden_attributes のキー（CSVヘッダーそのまま）を japaneseNames の
// キー形式に正規化する。14シーズン分の実データには同じ能力値でも次のような表記ゆれがあり、
// これらを1つのキーに集約するための変換を行う。
//
//  1. "stats/<name>/value" / "stats/<name>/diff"（2024年のみの入れ子形式）から <name> を取り出す
//  2. 末尾の "Rating"（2021年形式。例: "SpeedRating"）を除去する
//  3. 大文字/小文字・空白・記号の違いを除去する（既存の normalizeHeader と同じ方式）
//  4. なお正規化だけでは一致しない表記（"Block" と "Blocking" の違い、typo、省略形など）は
//     aliasKeys で読み替える
//
// isDiff は 2024年の "/diff" 列（前回更新からの変化量）だった場合に true になる。
func canonicalStatKey(raw string) (key string, isDiff bool) {
	if m := nestedStatPattern.FindStringSubmatch(raw); m != nil {
		return resolveAlias(normalizeHeader(m[1])), m[2] == "diff"
	}
	n := normalizeHeader(raw)
	if after, ok := strings.CutSuffix(n, "rating"); ok && after != "" {
		n = after
	}
	return resolveAlias(n), false
}

func resolveAlias(key string) string {
	if alias, ok := aliasKeys[key]; ok {
		return alias
	}
	return key
}

// aliasKeys は概念としては同じだが正規化だけでは一致しない表記を、主キー（japaneseNames の
// キー）に読み替えるための対応表。次のような原因による表記ゆれを吸収する。
//   - "Block" と "Blocking" のような言い回しの違い（impactBlock / impactBlocking）
//   - Madden 側の仕様変更で項目名が変わったもの（Footwork → Finesse, Strength → Power。
//     2017年以前の Pass/Run Block Footwork・Strength は、現行の Finesse・Power に相当する）
//   - 実データ中の typo（Finnesse Moves, Stength, Deep Throw Accruacy など）
//   - 年によって短縮・分割されている表記（Catch → Catching, Deep Accuracy → ThrowAccuracyDeep）
var aliasKeys = map[string]string{
	// 言い回しの違い（Block / Blocking）
	"impactblock":  "impactblocking",
	"leadblocking": "leadblock",
	"passblocking": "passblock",
	"runblocking":  "runblock",
	"mancover":     "mancoverage",
	"zonecover":    "zonecoverage",
	// Madden 側の仕様変更（Footwork→Finesse、Strength→Power）
	"passblockfootwork": "passblockfinesse",
	"passblockstrength": "passblockpower",
	"runblockfootwork":  "runblockfinesse",
	"runblockstrength":  "runblockpower",
	// 短縮形・別名
	"ballcarriervision": "bcvision",
	"catch":             "catching",
	"return":            "kickreturn",
	"playrec":           "playrecognition",
	"powermove":         "powermoves",
	"experience":        "yearspro",
	"primarykey":        "id",
	"playerportrait":    "playerimage",
	"portraitid":        "playerimage",
	"plyrbirthdate":     "birthdate",
	"handed":            "handedness",
	"plyrhandedness":    "handedness",
	"playerhandedness":  "handedness",
	"playerhandness":    "handedness", // typo: "Handness"（"ed" 抜け）
	// Throw/Route の「総合」列しか無かった年の列を、短距離/中距離/長距離の表記に統合
	"shortthrowaccuracy":  "throwaccuracyshort",
	"shortaccuracy":       "throwaccuracyshort",
	"mediumthrowaccuracy": "throwaccuracymid",
	"middleaccuracy":      "throwaccuracymid",
	"throwaccuracymiddle": "throwaccuracymid",
	"throwaccuracymed":    "throwaccuracymid",
	"deepthrowaccuracy":   "throwaccuracydeep",
	"deepthrowaccruacy":   "throwaccuracydeep", // typo: "Accruacy"
	"deepaccuracy":        "throwaccuracydeep",
	"throwonrun":          "throwontherun",
	"shortrouteruning":    "shortrouterunning", //nolint:misspell // 実データの typo "Runing"（n 抜け）をそのまま再現している
	// typo
	"finnessemoves":     "finessemoves",     // typo: "Finnesse"（n 過多）
	"finesseemove":      "finessemoves",     // typo: "Finessee" + 単数形
	"spectactularcatch": "spectacularcatch", // typo: "Spectactular"
	"stength":           "strength",         // typo: "Stength"（r 抜け）
}

// japaneseNames は canonicalStatKey で正規化したキー → 日本語名。
var japaneseNames = map[string]string{
	// ---- 能力値（フィジカル）----
	"speed":        "スピード",
	"acceleration": "加速力",
	"agility":      "アジリティ",
	"awareness":    "アウェアネス",
	"strength":     "パワー",
	"stamina":      "スタミナ",
	"injury":       "耐久力",
	"toughness":    "タフネス",
	"jumping":      "ジャンプ力",

	// ---- 能力値（ボールキャリア）----
	"elusiveness":       "回避力",
	"bcvision":          "ビジョン",
	"catching":          "キャッチ",
	"catchintraffic":    "混雑下でのキャッチ",
	"carrying":          "ボールキャリング",
	"breaktackle":       "タックルブレイク",
	"breaksack":         "サック回避",
	"trucking":          "トラッキング",
	"stiffarm":          "スティフアーム",
	"spinmove":          "スピンムーブ",
	"jukemove":          "ジュークムーブ",
	"changeofdirection": "方向転換",
	"spectacularcatch":  "スペクタキュラーキャッチ",

	// ---- 能力値（ディフェンス）----
	"release":         "リリース",
	"press":           "プレスカバー",
	"mancoverage":     "マンカバー",
	"zonecoverage":    "ゾーンカバー",
	"playrecognition": "プレー読み",
	"blockshedding":   "ブロック解除",
	"tackle":          "タックル",
	"hitpower":        "ヒットパワー",
	"powermoves":      "パワームーブ",
	"finessemoves":    "フィネスムーブ",
	"pursuit":         "パーシュート",

	// ---- 能力値（オフェンスライン）----
	"impactblocking":   "インパクトブロック",
	"leadblock":        "リードブロック",
	"runblock":         "ランブロック",
	"runblockpower":    "ランブロックパワー",
	"runblockfinesse":  "ランブロックフィネス",
	"passblock":        "パスブロック",
	"passblockpower":   "パスブロックパワー",
	"passblockfinesse": "パスブロックフィネス",

	// ---- 能力値（キック）----
	"kickpower":    "キックパワー",
	"kickaccuracy": "キックアキュラシー",
	"kickreturn":   "キックリターン",
	"runningstyle": "ランニングスタイル",

	// ---- 能力値（パス）----
	"throwpower":         "スロー力",
	"throwaccuracy":      "スローアキュラシー（総合）",
	"throwaccuracyshort": "スローアキュラシー（ショート）",
	"throwaccuracymid":   "スローアキュラシー（ミドル）",
	"throwaccuracydeep":  "スローアキュラシー（ディープ）",
	"throwontherun":      "ラン中のスロー精度",
	"throwunderpressure": "プレッシャー下でのスロー精度",
	"playaction":         "プレーアクション",

	// ---- 能力値（レシーバーのルート）----
	"routerunning":       "ルートランニング（総合）",
	"shortrouterunning":  "ルートランニング（ショート）",
	"mediumrouterunning": "ルートランニング（ミドル）",
	"deeprouterunning":   "ルートランニング（ディープ）",

	// ---- 総合値 ----
	"overall": "オーバーオール（総合値）",

	// ---- 選手の基本情報 ----
	"age":        "年齢",
	"archetype":  "アーキタイプ（選手タイプ）",
	"college":    "大学",
	"conference": "カンファレンス",
	"division":   "ディビジョン",
	"yearspro":   "プロ経験年数",
	"height":     "身長",
	"weight":     "体重",
	"birthdate":  "生年月日",
	"handedness": "利き手",
	"hometown":   "出身地",
	"homestate":  "出身州",
	"firstname":  "名",
	"lastname":   "姓",
	"position":   "ポジション",

	// ---- 契約・画像・Madden内部情報 ----
	"totalsalary":  "年俸総額",
	"signingbonus": "契約ボーナス",
	"playerimage":  "選手画像",
	"teamlogo":     "チームロゴ",
	"id":           "選手の内部ID（Madden側）",
	"teamid":       "チームの内部ID（Madden側）",
}

// JapaneseLabel は madden_attributes のキー（CSVヘッダーそのまま）に対応する日本語名を返す。
// 見つからなければ ok=false。
// 2024年形式の "/diff" 列（前回更新からの変化量）は、対応する日本語名に「（変化量）」を付加する。
func JapaneseLabel(rawHeader string) (label string, ok bool) {
	key, isDiff := canonicalStatKey(rawHeader)
	name, ok := japaneseNames[key]
	if !ok {
		return "", false
	}
	if isDiff {
		return name + "（変化量）", true
	}
	return name, true
}
