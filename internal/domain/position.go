package domain

// ポジションのカテゴリ（positions.category / pf_relations.position_category）。
// Laravel の config/const.php の Category と同じ値。
const (
	CategoryQB      int32 = 1
	CategoryHB      int32 = 2
	CategoryFB      int32 = 3
	CategoryWR      int32 = 4
	CategoryTE      int32 = 5
	CategoryRG      int32 = 6
	CategoryLG      int32 = 7
	CategoryROT     int32 = 8
	CategoryLOT     int32 = 9
	CategoryC       int32 = 10
	CategoryRDT     int32 = 11
	CategoryLDT     int32 = 12
	CategoryNT      int32 = 13
	CategoryRDE     int32 = 14
	CategoryLDE     int32 = 15
	CategoryMLB     int32 = 16
	CategoryRLB     int32 = 17
	CategoryLLB     int32 = 18
	CategoryRCB     int32 = 19
	CategoryLCB     int32 = 20
	CategorySS      int32 = 21
	CategoryFS      int32 = 22
	CategoryK       int32 = 23
	CategoryP       int32 = 24
	CategoryLS      int32 = 25
	CategoryNothing int32 = 99
)

// ポジションの抽象カテゴリ（positions.abstract_category / pf_relations.abstract_category）。
// Laravel の config/const.php の AbstCate と同じ値。
const (
	AbstractCategoryQB      int32 = 1
	AbstractCategoryRB      int32 = 2
	AbstractCategoryWR      int32 = 3
	AbstractCategoryTE      int32 = 4
	AbstractCategoryOL      int32 = 5
	AbstractCategoryDL      int32 = 6
	AbstractCategoryLB      int32 = 7
	AbstractCategoryDB      int32 = 8
	AbstractCategoryK       int32 = 9
	AbstractCategoryLS      int32 = 10
	AbstractCategoryNothing int32 = 99
)

// PositionNoData はポジション情報が無い選手に割り当てるポジション名。
const PositionNoData = "No Data"

// PositionDetail はポジション1件の情報。API（/api/v1/rosters）で返す読み取り用エンティティ。
// positions テーブルの id を持つ点がマスタ定義の Position（position_data.go）と異なる。
type PositionDetail struct {
	ID               int32
	AbstractCategory int32
	Category         int32
	Name             string
	Odflg            bool
}
