// Code generated from the Laravel seeders (database/seeders/FormationsTableSeeder.php) by a one-off script; hand edits are fine.

package seed

// formation はフォーメーションの初期データ 1 件。Key は Laravel の config/const.php の Formationid のキーで、
// pf_relations から参照するために使う（テーブルには保存しない）。
type formation struct {
	Key   string
	Odflg bool
	Name  string
}

// formations は formations テーブルの初期データ。順序は Laravel の FormationsTableSeeder（= Formationid の値順）と同じ。
var formations = []formation{

	// ---- Offence Category ----
	// Tフォーメーション
	{Key: "T", Odflg: true, Name: "T formation"},
	// Iフォーメーション
	{Key: "I", Odflg: true, Name: "I formation"},
	// シングルバック
	{Key: "SSB", Odflg: true, Name: "Single set back"},
	// スプレッドフォーメーション
	{Key: "SF", Odflg: true, Name: "Spread formation"},
	// プロセット(スプリットバック) FB & HB
	{Key: "PSFH", Odflg: true, Name: "Pro set(Split back) FB & HB"},
	// プロセット(スプリットバック) HB & HB
	{Key: "PSHH", Odflg: true, Name: "Pro set(Split back) HB & HB"},
	// ショットガン
	{Key: "SG", Odflg: true, Name: "Shotgun"},
	// ピストルフォーメーション
	{Key: "PF", Odflg: true, Name: "Pistol formation"},
	// ウィッシュボーン
	{Key: "WB", Odflg: true, Name: "Wishbone"},
	// フレックスボーン (WRが4人バージョン)
	{Key: "FBW4", Odflg: true, Name: "Flexbone WR4"},
	// フレックスボーン (WRが2人・TEが2人バージョン)
	{Key: "FBWT", Odflg: true, Name: "Flexbone WR2 & TE2"},
	// ウィングT
	{Key: "WT", Odflg: true, Name: "Wing T"},
	// エンプティバックフィールド
	{Key: "EB", Odflg: true, Name: "Empty backfield"},
	// シングルウィング
	{Key: "SW", Odflg: true, Name: "Single wing"},
	// ワイルドキャットフォーメーション
	{Key: "WF", Odflg: true, Name: "Wildcat formation"},

	// ---- Defence Category ----
	// 3-4
	{Key: "3-4", Odflg: false, Name: "3-4"},
	// 4-3
	{Key: "4-3", Odflg: false, Name: "4-3"},
	// 4-4
	{Key: "4-4", Odflg: false, Name: "4-4"},
	// 5-2
	{Key: "5-2", Odflg: false, Name: "5-2"},
	// Nickel(ニッケル) ※LCBが多い場合
	{Key: "Nickel_LCB", Odflg: false, Name: "Nickel(LCB)"},
	// Nickel(ニッケル) ※RCBが多い場合
	// Laravel 側では "Nickel(LCB)" と重複していたため、キー Nickel_RCB に合わせて修正
	{Key: "Nickel_RCB", Odflg: false, Name: "Nickel(RCB)"},
	// Dime(ダイム)
	{Key: "Dime", Odflg: false, Name: "Dime"},
	// 46(フォーシックス)
	{Key: "46", Odflg: false, Name: "46"},
	// 1-6-4
	{Key: "1-6-4", Odflg: false, Name: "1-6-4"},
	// 2-4-5
	{Key: "2-4-5", Odflg: false, Name: "2-4-5"},
	// 3-7-1
	{Key: "3-7-1", Odflg: false, Name: "3-7-1"},
	// 3-6-2
	{Key: "3-6-2", Odflg: false, Name: "3-6-2"},
	// 4-6-1
	{Key: "4-6-1", Odflg: false, Name: "4-6-1"},
	// 4-5-2
	{Key: "4-5-2", Odflg: false, Name: "4-5-2"},
	// 5-5-1
	{Key: "5-5-1", Odflg: false, Name: "5-5-1"},
	// 5-4-2
	{Key: "5-4-2", Odflg: false, Name: "5-4-2"},
	// 6-4-1
	{Key: "6-4-1", Odflg: false, Name: "6-4-1"},
	// 6-3-2
	{Key: "6-3-2", Odflg: false, Name: "6-3-2"},
	// 3-3-5
	{Key: "3-3-5", Odflg: false, Name: "3-3-5"},
	// シングルバック TE
	{Key: "SSB_TE", Odflg: true, Name: "Single set back TE"},
	// Iフォーメーション TE
	{Key: "I_TE", Odflg: true, Name: "I formation TE"},
}
