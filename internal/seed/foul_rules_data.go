// Code generated from the Laravel seeders (database/seeders/FoulRulesTableSeeder.php) by a one-off script; hand edits are fine.

package seed

import (
	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// foulRules は foul_rules テーブルの初期データ。順序は Laravel の FoulRulesTableSeeder と同じ。
var foulRules = []sqlc.InsertFoulRulesParams{

	// ---- Offence Defence Common Category ----
	// Illegal Use Of Hand
	{
		EnglishName:   "Illegal Use Of Hand",
		JapaneseName:  "不正な手の使用",
		Description:   "相手選手を対するに当たり、手や腕の使い方のルールが決まっており、それに反する行為をすること。攻守共通してフェイスマスクを押す行為も含まれるが、ボールキャリアーに限り対象外となる。",
		StatusType:    domain.FoulStatusCommon,
		YardInfo:      10,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Personal Foul
	{
		EnglishName:   "Personal Foul",
		JapaneseName:  "パーソナル・ファウル",
		Description:   "怪我を招く危険な行為は、故意でも偶然でも、パーソナル・ファウルと呼ばれる。守備側のパーソナル・ファウルの場合、攻撃側は自動的にファーストダウンが与えられる。\nまた、危険度が高かったり、悪質と判断したりした場合は、反則を犯した選手を「資格没収（退場）」の処分が科される（相手チームの辞退、相殺を問わない）。",
		StatusType:    domain.FoulStatusCommon,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       true,
	},
	// Facemask
	{
		EnglishName:   "Facemask",
		JapaneseName:  "フェイス・マスク",
		Description:   "フェイス・マスクを掴んで、引っ張ったり、捻ったりする行為。",
		StatusType:    domain.FoulStatusCommon,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Unsportsmanlike Conduct
	{
		EnglishName:   "Unsportsmanlike Conduct",
		JapaneseName:  "スポーツマンらしからぬ行為",
		Description:   "審判、相手選手、観客への暴言、侮蔑行為。特に悪質なものはトーンティング（taunting）とも呼ばれる。\n得点やビッグプレー後の過度のセレブレーション行為も対象である。\n守備側の選手によるものの場合は、攻撃側にファーストダウンが与えられる。",
		StatusType:    domain.FoulStatusCommon,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},

	// ---- Offence Category ----
	// Delay Of Game
	{
		EnglishName:   "Delay Of Game",
		JapaneseName:  "ゲームの遅延",
		Description:   "審判がレディー・フォー・プレーを宣告してから設定されていた40秒また25秒以内にプレーを開始しなかった。またタイムアウトの権利が残っていないのにタイムアウトを要求したこと。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// False Start
	{
		EnglishName:   "False Start",
		JapaneseName:  "フォルス・スタート",
		Description:   "ボールがスナップされる前に、攻撃側の選手が攻撃と勘違いされるような動きをすること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Motion
	{
		EnglishName:   "Illegal Motion",
		JapaneseName:  "イリーガル・モーション",
		Description:   "攻撃側は、1人だけスナップ直前でも動いていることが許される。ただし、次の3つのルールを守る必要があり、満たしていない場合はイリーガル・モーション。\n(1) 攻撃側の選手全員が、1秒以上静止したのちに動くこと。\n(2) 動く選手はバックの選手であること。\n(3) 相手ゴールライン方向（前方）へ動かないこと。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Shift
	{
		EnglishName:   "Illegal Shift",
		JapaneseName:  "イリーガル・シフト",
		Description:   "攻撃側は、2人以上の選手が動いたり、1人以上が前方に動作したりして、ポジションや姿勢を変更することが許される。ただし、その変更後、全員が1秒以上静止する必要がある。静止しないままスナップした場合、イリーガル・シフトとなる。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Formation
	{
		EnglishName:   "Illegal Formation",
		JapaneseName:  "不正なフォーメーション",
		Description:   "攻撃側はスナップするとき、特定の位置に選手がいる必要があったり、逆にいてはならなかったり、フォーメーションに関するルールがある。このルールに反するとイリーガル・フォーメーション。\n具体的には次の3点が当てはまる。\n(1) スクリメージラインから1ヤード以内の選手（ライン上の選手）が7人以上いること。\n(2) NFLでは、ラインの内側が無資格番号(50-79)の選手であること。\n(3) NFLでは、ラインの両端やバックが有資格番号(1-49、80-89)の選手であること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Substitution
	{
		EnglishName:   "Illegal Substitution",
		JapaneseName:  "交代違反",
		Description:   "12人以上の選手がハドルしたり、フォーメーションについたりすること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Ineligible Receiver Downfield on Pass
	{
		EnglishName:   "Ineligible Receiver Downfield on Pass",
		JapaneseName:  "無資格レシーバーのダウンフィールドへの侵入",
		Description:   "パスプレーで、番号による無資格選手またはポジションによる無資格の選手（主にオフェンスラインの選手）が、パスが投げられる前にスクリメージラインより3ヤード以上（NFLでは1ヤード以上）前に出ること。\nQBをプロテクションする(守る)はずのラインの選手が前にでるとランプレーであると勘違いし、守備側が不利になってしまうため、反則として扱われる。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Forward Pass
	{
		EnglishName:   "Illegal Forward Pass",
		JapaneseName:  "不正なフォワードパス",
		Description:   "ボールキャリアがスクリメージラインを超えてから前方にパスした。または1回のプレーで2回以上前方にパスした。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      5,
		LossofdownFlg: true,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Pass Interference
	{
		EnglishName:   "Pass Interference",
		JapaneseName:  "パス・インターフェース",
		Description:   "攻撃選手による守備選手へのパス・キャッチの妨害をすること。具体的には次の2点が当てはまる。\n(1) 守備選手を押すなど、ランプレーのためのブロックと勘違いさせる行為をした後に、パスコースにでて、その後パスが投げられたなど（※パスの方向は無関係でも当てはまる）。\n(2) パスカバーしている守備選手を押して間合いを取って、パスキャッチを試みるなど、守備選手のボールをキャッチする機会を妨害すること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      10,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Intentional Grounding
	{
		EnglishName:   "Intentional Grounding",
		JapaneseName:  "インテンショナル・グラウンディング",
		Description:   "ロスを逃れるため、有資格レシーバーがいない地域に 故意にパスを投げること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      10,
		LossofdownFlg: true,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Holding
	{
		EnglishName:   "Holding",
		JapaneseName:  "ホールディング",
		Description:   "パス・プロテクションやラン・ブロックで、相手の腕や足など身体やジャージを掴んだり、相手に巻き付いたりすること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      10,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Clipping
	{
		EnglishName:   "Clipping",
		JapaneseName:  "クリッピング",
		Description:   "最初の接触が背後で、かつ、腰を含めて腰から下へのブロックすること。\nただし、スナップしたときに決められた領域にいた選手が、その領域内でのブロックについては、膝を含めて膝から下へのブロックでなければ、反則ではない。\n※NFLでは、クローズライン・プレー(Close-line Play)と呼び、両タックルの間で、スクリメージライン前後3ヤードの長方形の領域がそこに当てはまる。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Block In the Back
	{
		EnglishName:   "Block In the Back",
		JapaneseName:  "背後のブロック",
		Description:   "相手を背後から腰から上に接触してブロックすること。\nただし、クリッピングと同様の領域でのブロックは反則ではない。また、明らかにブロックしようとしてから、相手選手が背中を向けたときも反則にならない。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      10,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Chop Block
	{
		EnglishName:   "Chop Block",
		JapaneseName:  "チョップ・ブロック",
		Description:   "2人の攻撃選手が1人の守備選手に対して、片方がロー、他方がハイの組み合わせによるブロックをすること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Tripping
	{
		EnglishName:   "Tripping",
		JapaneseName:  "トリッピング",
		Description:   "足または脚を使って相手選手へブロックすること。",
		StatusType:    domain.FoulStatusOffence,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},

	// ---- Defence Category ----
	// Offside
	{
		EnglishName:   "Offside",
		JapaneseName:  "オフサイド",
		Description:   "スナップ前に守備選手が不正な位置にいること。具体的には次の場合が該当する。\n(1) スナップしたときに守備選手がニュートラルゾーン内か、ニュートラルゾーンに越えた位置にいる。スナップ前に戻っていれば反則ではない。\n(2) スナップする前に守備選手が、ニュートラルゾーンを越えて攻撃選手に接触する。\n(3) スナップ前のボールを触る。\n(4) スナップ前に、相手選手に接触しなくても、バックに向かって突進する。\n(5) スナップ前に、オフェンスラインを脅して、反射的な動き（フォルス・スタート）を誘発する。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Encroachment
	{
		EnglishName:   "Encroachment",
		JapaneseName:  "エンクローチメント",
		Description:   "スナップする前に守備選手が、ニュートラルゾーンを越えて攻撃選手に接触すること。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Neutral Zone Infraction
	{
		EnglishName:   "Neutral Zone Infraction",
		JapaneseName:  "ニュートラルゾーン・インフラクション",
		Description:   "スナップする前に、ニュートラル・ゾーンに侵入し、攻撃選手の反射的な動き（フォルス・スタート）を誘発すること。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Pass Interference
	{
		EnglishName:   "Pass Interference",
		JapaneseName:  "パス・インターフェース",
		Description:   "守備選手による攻撃選手へのパス・キャッチの妨害をすること。次のような場合、反則と判断されることが多い。\n(1) 自身はボールをキャッチする構えなく、キャッチ（インターセプト）しようとしている相手選手に接触した。\n(2) 自身はボールに向かわずボールを追いかけている相手選手の動き（走路）を邪魔するために接触した。\n(3) パスキャッチしようとしている攻撃選手の肩や胸などに片手で接触して、他方の手でパスをブロックした。\nこれらの行為より前に、パスされたボールがタッチされている場合は、反則にならない。\nまた妨害がなかったとしても、明らかにキャッチできなかったと判断できると場合は、反則にならない。\nNFLでは、距離に関係なく反則地点でファーストダウンが与えられる。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      20,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Contact
	{
		EnglishName:   "Illegal Contact",
		JapaneseName:  "イリーガル・コンタクト",
		Description:   "NFL独自のルール。スクリメージ・ラインから5ヤード以上越えた地点で攻撃側のレシーバーに接触すること。\nボールがポケットから出た場合は、適用外となる。\nまた、攻撃選手から接触を受けて、それに対して自身の身体を守るための接触は認められる。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Holding
	{
		EnglishName:   "Holding",
		JapaneseName:  "ホールディング",
		Description:   "ボールを持っていない選手に対して、タックルしたり、つかんだり、その他妨害したりすること。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Roughing the Passer
	{
		EnglishName:   "Roughing the Passer",
		JapaneseName:  "ラフィング・ザ・パサー",
		Description:   "明らかにパスを投げた後のパサー（クォーターバック）を突き当たったり、投げつけたりの乱暴行為をすること。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Horse Collar Tackle
	{
		EnglishName:   "Horse Collar Tackle",
		JapaneseName:  "ホースカラータックル",
		Description:   "ボールキャリアーの襟首を掴んだタックル。乗馬で手綱を引いて止めるような形からの命名であり、首や腰が大きく反ってしまい危険が伴うため、反則として扱われる。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Formation
	{
		EnglishName:   "Illegal Formation",
		JapaneseName:  "不正なフォーメーション",
		Description:   "キックプレー時においてLOSの左右どちらかに6人以上が位置していること。",
		StatusType:    domain.FoulStatusDefence,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},

	// ---- Kick Category ----
	// Illegal Formation
	{
		EnglishName:   "Illegal Formation",
		JapaneseName:  "不正なフォーメーション",
		Description:   "フリーキック（キックオフ）をするチームは、次の2点を守らないといけない。\n(1) キッカー以外の選手はボールから5ヤード以内にいること。\n(2) キックしたとき、キッカーの両側にそれぞれ4人以上いること。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Offside
	{
		EnglishName:   "Offside",
		JapaneseName:  "オフサイド",
		Description:   "フリーキック（キックオフ）をする前に、キッカー、ホルダー以外の選手がボールより前に出ること。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Kickoff Out of Bounds
	{
		EnglishName:   "Kickoff Out of Bounds",
		JapaneseName:  "フリーキックのアウト・オブ・バウンズ",
		Description:   "フリーキック（キックオフ）されたボールが、レシーブ・チームの選手に触れることなく、ゴールラインより手前でサイドラインを割ること。\nパントキックではないキックオフのキックで、このペナルティが該当する。\nレシーブ・チーム陣40ヤードでファーストダウンが開始される。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      20,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Illegal Touch
	{
		EnglishName:   "Illegal Touch",
		JapaneseName:  "不正なタッチ",
		Description:   "フリーキック（キックオフ）されたボールは、キックした位置から10ヤード進むか、レシーブ・チームの選手に触れるまで、キック・チームの選手が触れてはならない。\n触れた地点から5ヤードレシービングチーム側に下がって、レシーブ・チームのファーストダウン。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Roughing the Kicker
	{
		EnglishName:   "Roughing the Kicker",
		JapaneseName:  "ラフィング・ザ・キッカー",
		Description:   "明らかに蹴った後のキッカーやパンターに突き当たったり、投げつけたりの乱暴行為をすること。\nまた、フィールドゴール・プレーのときホルダー(ボールをセットする人)に対しても同様の保護規定がある。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      15,
		LossofdownFlg: false,
		AfFlg:         true,
		ClockFlg:      false,
		ExitFlg:       false,
	},
	// Running Into the Kicker
	{
		EnglishName:   "Running Into the Kicker",
		JapaneseName:  "ランニング・イントゥ・ザ・キッカー",
		Description:   "明らかに蹴った後のキッカーやパンターに対して、乱暴とは言えないほどの接触があった場合にコールされる。ラフィング・ザ・キッカーほどではないが、キッカーが蹴り上げた足先に触ったときなど、多少の危険が伴うため、反則として扱われる。",
		StatusType:    domain.FoulStatusKick,
		YardInfo:      5,
		LossofdownFlg: false,
		AfFlg:         false,
		ClockFlg:      false,
		ExitFlg:       false,
	},
}
