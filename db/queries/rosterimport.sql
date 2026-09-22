-- cmd/importrosters（outputs/csv/*.csv から players / rosters を生成するコマンド）専用のクエリ。

-- name: GetPlayerByIdentity :one
SELECT * FROM players
WHERE firstname = $1 AND lastname = $2 AND birthday = $3;

-- name: CreatePlayer :one
INSERT INTO players (firstname, lastname, birthday, height, weight, college,
                      drafted_team, drafted_round, drafted_rank, drafted_year)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListPositions :many
SELECT * FROM positions
ORDER BY id;

-- name: UpsertRoster :one
INSERT INTO rosters (season, team_id, player_id, position_id, number, experience)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT ON CONSTRAINT rosters_season_team_player_key
DO UPDATE SET position_id = EXCLUDED.position_id,
              number       = EXCLUDED.number,
              experience   = EXCLUDED.experience,
              updated_at   = now()
RETURNING *;
