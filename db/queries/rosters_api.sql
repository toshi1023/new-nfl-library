-- GET /api/v1/rosters（season・team_id で絞り込んでロスターを返す API）専用のクエリ。

-- name: ListRostersBySeasonTeam :many
SELECT
    rosters.id,
    rosters.season,
    rosters.team_id,
    rosters.player_id,
    rosters.position_id,
    rosters.number,
    rosters.rating,
    rosters.experience,
    rosters.madden_attributes,
    players.id AS player_table_id,
    players.firstname,
    players.lastname,
    players.birthday,
    players.height,
    players.weight,
    players.college,
    players.drafted_team,
    players.drafted_round,
    players.drafted_rank,
    players.drafted_year,
    players.image_file,
    positions.id AS position_table_id,
    positions.abstract_category,
    positions.category,
    positions.name AS position_name,
    positions.odflg
FROM rosters
JOIN players ON players.id = rosters.player_id
JOIN positions ON positions.id = rosters.position_id
WHERE rosters.season = $1 AND rosters.team_id = $2
  AND (
    cardinality(sqlc.arg('names')::text[]) = 0
    OR EXISTS (
      SELECT 1 FROM unnest(sqlc.arg('names')::text[]) AS term
      WHERE players.firstname ILIKE '%' || term || '%' ESCAPE '\'
         OR players.lastname ILIKE '%' || term || '%' ESCAPE '\'
    )
  )
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'rating' AND sqlc.arg('sort_desc')::bool THEN rosters.rating END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'rating' AND NOT sqlc.arg('sort_desc')::bool THEN rosters.rating END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'number' AND sqlc.arg('sort_desc')::bool THEN rosters.number END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'number' AND NOT sqlc.arg('sort_desc')::bool THEN rosters.number END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'experience' AND sqlc.arg('sort_desc')::bool THEN rosters.experience END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'experience' AND NOT sqlc.arg('sort_desc')::bool THEN rosters.experience END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_desc')::bool THEN players.lastname END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND NOT sqlc.arg('sort_desc')::bool THEN players.lastname END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_desc')::bool THEN players.firstname END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND NOT sqlc.arg('sort_desc')::bool THEN players.firstname END ASC,
  players.lastname,
  players.firstname;
