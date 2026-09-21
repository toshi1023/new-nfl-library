// Code generated from the Laravel seeder (database/seeders/PositionsTableSeeder.php) by a one-off script; hand edits are fine.

package domain

// Position はポジションマスタの 1 件。internal/seed の DB 投入と、スクレイプしたポジション名から
// カテゴリを引く用途（cmd/scrapestarters）の両方から参照される唯一の定義元。
type Position struct {
	AbstractCategory int32
	Category         int32
	Name             string
	// Odflg は攻守フラグ（true: Offence, false: Defence）。positions テーブルの管理用フラグであり、
	// K/P/LS のようなスペシャルチームも true になる。攻守判定には使わず IsOffenceCategory を使うこと。
	Odflg bool
}

// allPositions はポジションマスタの全件。順序は Laravel の PositionsTableSeeder と同じ。
var allPositions = []Position{

	// ---- Offence Category ----
	// QB
	{AbstractCategory: AbstractCategoryQB, Category: CategoryQB, Name: "QB", Odflg: true},
	// RB
	{AbstractCategory: AbstractCategoryRB, Category: CategoryHB, Name: "RB", Odflg: true},
	// HB
	{AbstractCategory: AbstractCategoryRB, Category: CategoryHB, Name: "HB", Odflg: true},
	// FB
	{AbstractCategory: AbstractCategoryRB, Category: CategoryFB, Name: "FB", Odflg: true},
	// WR
	{AbstractCategory: AbstractCategoryWR, Category: CategoryWR, Name: "WR", Odflg: true},
	// WR/RB([saints]Ty Montgomery)
	{AbstractCategory: AbstractCategoryWR, Category: CategoryWR, Name: "WR/RB", Odflg: true},
	// WR/CB([lions]Jamal Agnew)
	{AbstractCategory: AbstractCategoryWR, Category: CategoryWR, Name: "WR/CB", Odflg: true},
	// PR([bears]Devin Hester)
	{AbstractCategory: AbstractCategoryWR, Category: CategoryWR, Name: "PR", Odflg: true},
	// TE
	{AbstractCategory: AbstractCategoryTE, Category: CategoryTE, Name: "TE", Odflg: true},
	// QB/TE([lions]Logan Thomas)
	{AbstractCategory: AbstractCategoryTE, Category: CategoryTE, Name: "QB/TE", Odflg: true},
	// RG
	{AbstractCategory: AbstractCategoryOL, Category: CategoryRG, Name: "RG", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryRG, Name: "RG/RT", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryRG, Name: "RG/C", Odflg: true},
	// G
	{AbstractCategory: AbstractCategoryOL, Category: CategoryRG, Name: "G", Odflg: true},
	// LG
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLG, Name: "LG", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLG, Name: "LG/LT", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLG, Name: "LG/C", Odflg: true},
	// OG
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLG, Name: "OG", Odflg: true},
	// RT
	{AbstractCategory: AbstractCategoryOL, Category: CategoryROT, Name: "RT", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryROT, Name: "RT/LT", Odflg: true},
	// OT
	{AbstractCategory: AbstractCategoryOL, Category: CategoryROT, Name: "OT", Odflg: true},
	// LT
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLOT, Name: "LT", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLOT, Name: "LT/RT", Odflg: true},
	// T
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLOT, Name: "T", Odflg: true},
	// G,T([patriots]Marshall Newhouse)
	{AbstractCategory: AbstractCategoryOL, Category: CategoryLOT, Name: "G,T", Odflg: true},
	// C
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "C", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "C/RG", Odflg: true},
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "C/LG", Odflg: true},
	// OL
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "OL", Odflg: true},
	// G/C
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "G/C", Odflg: true},
	// G,C
	{AbstractCategory: AbstractCategoryOL, Category: CategoryC, Name: "G,C", Odflg: true},

	// ---- Defence Category ----
	// RDT
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDT, Name: "RDT", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDT, Name: "RDT/RDE", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDT, Name: "RDT/LDT", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDT, Name: "RDT/LDE", Odflg: false},
	// DT
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDT, Name: "DT", Odflg: false},
	// LDT
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDT, Name: "LDT", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDT, Name: "LDT/LDE", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDT, Name: "LDT/RDT", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDT, Name: "LDT/RDE", Odflg: false},
	// DL
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDT, Name: "DL", Odflg: false},
	// NT
	{AbstractCategory: AbstractCategoryDL, Category: CategoryNT, Name: "NT", Odflg: false},
	// RDE
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDE, Name: "RDE", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDE, Name: "RDE/LDE", Odflg: false},
	// RE
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDE, Name: "RE", Odflg: false},
	// EDGE([49ers]Nick Bosa)
	{AbstractCategory: AbstractCategoryDL, Category: CategoryRDE, Name: "EDGE", Odflg: false},
	// LDE
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDE, Name: "LDE", Odflg: false},
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDE, Name: "LDE/RDE", Odflg: false},
	// LE
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDE, Name: "LE", Odflg: false},
	// DE
	{AbstractCategory: AbstractCategoryDL, Category: CategoryLDE, Name: "DE", Odflg: false},
	// MLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryMLB, Name: "MLB", Odflg: false},
	// IL
	{AbstractCategory: AbstractCategoryLB, Category: CategoryMLB, Name: "IL", Odflg: false},
	// ILB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryMLB, Name: "ILB", Odflg: false},
	// RILB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryMLB, Name: "RILB", Odflg: false},
	// LILB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryMLB, Name: "LILB", Odflg: false},
	// RLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "RLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "RLB/LLB", Odflg: false},
	// ROLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "ROLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "ROLB/LOLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "ROLB/LILB", Odflg: false},
	// OLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "OLB", Odflg: false},
	// DT/LB([bills]Lorenzo Alexander)
	{AbstractCategory: AbstractCategoryLB, Category: CategoryRLB, Name: "DT/LB", Odflg: false},
	// LLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LLB/RLB", Odflg: false},
	// LOLB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LOLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LOLB/ROLB", Odflg: false},
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LOLB/RILB", Odflg: false},
	// LB
	{AbstractCategory: AbstractCategoryLB, Category: CategoryLLB, Name: "LB", Odflg: false},
	// RCB
	{AbstractCategory: AbstractCategoryDB, Category: CategoryRCB, Name: "RCB", Odflg: false},
	{AbstractCategory: AbstractCategoryDB, Category: CategoryRCB, Name: "RCB/LCB", Odflg: false},
	// CB
	{AbstractCategory: AbstractCategoryDB, Category: CategoryRCB, Name: "CB", Odflg: false},
	// LCB
	{AbstractCategory: AbstractCategoryDB, Category: CategoryLCB, Name: "LCB", Odflg: false},
	{AbstractCategory: AbstractCategoryDB, Category: CategoryLCB, Name: "LCB/RCB", Odflg: false},
	// DB
	{AbstractCategory: AbstractCategoryDB, Category: CategoryLCB, Name: "DB", Odflg: false},
	// SS
	{AbstractCategory: AbstractCategoryDB, Category: CategorySS, Name: "SS", Odflg: false},
	{AbstractCategory: AbstractCategoryDB, Category: CategorySS, Name: "SS/FS", Odflg: false},
	// S
	{AbstractCategory: AbstractCategoryDB, Category: CategorySS, Name: "S", Odflg: false},
	// FS
	{AbstractCategory: AbstractCategoryDB, Category: CategoryFS, Name: "FS", Odflg: false},
	{AbstractCategory: AbstractCategoryDB, Category: CategoryFS, Name: "FS/SS", Odflg: false},

	// ---- Special Category ----
	// K
	{AbstractCategory: AbstractCategoryK, Category: CategoryK, Name: "K", Odflg: true},
	// P
	{AbstractCategory: AbstractCategoryK, Category: CategoryP, Name: "P", Odflg: true},
	// LS
	{AbstractCategory: AbstractCategoryLS, Category: CategoryLS, Name: "LS", Odflg: true},
	// LS,TE
	{AbstractCategory: AbstractCategoryLS, Category: CategoryLS, Name: "LS,TE", Odflg: true},
	// ポジション情報が無い場合
	{AbstractCategory: AbstractCategoryNothing, Category: CategoryNothing, Name: PositionNoData, Odflg: true},
}

// AllPositions はポジションマスタの全件を返す。
func AllPositions() []Position {
	return allPositions
}

// PositionCategory は名前からカテゴリを引く。見つからなければ ok=false。
func PositionCategory(name string) (category int32, ok bool) {
	for _, p := range allPositions {
		if p.Name == name {
			return p.Category, true
		}
	}
	return 0, false
}

// IsOffenceCategory はカテゴリが攻撃側かどうかを返す。CategoryC(10) までが攻撃、CategoryRDT(11) 以降が守備。
// Laravel の ScrapeStarters（category < 11）と同じ境界値で、Position.Odflg とは意味が異なる点に注意
// （Odflg は K/P/LS のようなスペシャルチームも true にする管理用フラグ）。
func IsOffenceCategory(category int32) bool {
	return category < CategoryRDT
}
