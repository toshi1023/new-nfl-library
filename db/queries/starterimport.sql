-- cmd/importdata（db/data の starter CSV を starters テーブルへ反映するコマンド）専用のクエリ。
-- roster の検索は maddenimport.sql の GetRosterByPlayerIdentity を共用する。

-- name: UpsertStarter :exec
INSERT INTO starters (season, odflg, roster_id)
VALUES ($1, $2, $3)
ON CONFLICT ON CONSTRAINT starters_roster_id_key
DO UPDATE SET season     = EXCLUDED.season,
              odflg      = EXCLUDED.odflg,
              updated_at = now();
