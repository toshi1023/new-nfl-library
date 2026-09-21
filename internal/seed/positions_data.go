package seed

import (
	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// positions は positions テーブルの初期データ。実体は domain.AllPositions()（Laravel の
// PositionsTableSeeder から生成した唯一のマスタ定義）で、ここでは sqlc の投入用パラメータに変換するだけ。
var positions = toInsertPositionsParams(domain.AllPositions())

func toInsertPositionsParams(defs []domain.Position) []sqlc.InsertPositionsParams {
	out := make([]sqlc.InsertPositionsParams, len(defs))
	for i, d := range defs {
		out[i] = sqlc.InsertPositionsParams{
			AbstractCategory: d.AbstractCategory,
			Category:         d.Category,
			Name:             d.Name,
			Odflg:            d.Odflg,
		}
	}
	return out
}
