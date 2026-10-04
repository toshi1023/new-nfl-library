-- cmd/importmadden（Madden NFL 評価CSVで rosters を更新するコマンド）専用のクエリ。

-- name: FindTeamByName :one
SELECT * FROM teams
WHERE name = $1 OR (city || ' ' || name) = $1;

-- name: GetRosterByPlayerIdentity :one
SELECT rosters.* FROM rosters
JOIN players ON players.id = rosters.player_id
WHERE rosters.season = $1 AND rosters.team_id = $2
  AND players.firstname = $3 AND players.lastname = $4;

-- name: UpdateRosterMaddenAttributes :exec
UPDATE rosters
SET position_id       = $1,
    rating             = $2,
    number             = $3,
    madden_attributes  = $4,
    updated_at         = now()
WHERE id = $5;
