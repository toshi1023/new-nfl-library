package repository

import "github.com/jackc/pgx/v5/pgtype"

// numericToFloat は pgtype.Numeric を *float64 に変換する。NULL（Valid: false）なら nil を返す。
func numericToFloat(n pgtype.Numeric) (*float64, error) {
	if !n.Valid {
		return nil, nil
	}
	f, err := n.Float64Value()
	if err != nil {
		return nil, err
	}
	v := f.Float64
	return &v, nil
}
