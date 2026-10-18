-- GET /api/v1/rosters（season・team_id で絞り込んでロスターを返す API）・
-- GET /api/v1/rosters/{id}（ロスター1件の詳細を返す API）専用のクエリ。

-- name: ListRostersBySeasonTeam :many
-- 一覧では madden_attributes は返さない（詳細 API の GetRosterDetailByID でのみ返す）。
SELECT
    rosters.id,
    rosters.season,
    rosters.team_id,
    rosters.player_id,
    rosters.position_id,
    rosters.number,
    rosters.rating,
    rosters.experience,
    sqlc.embed(players),
    sqlc.embed(positions)
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

-- name: GetRosterDetailByID :one
-- starters は該当しない roster もあるため LEFT JOIN（スターターでなければ starter_id が NULL）。
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
    sqlc.embed(players),
    sqlc.embed(positions),
    sqlc.embed(teams),
    starters.id AS starter_id,
    starters.odflg AS starter_odflg
FROM rosters
JOIN players ON players.id = rosters.player_id
JOIN positions ON positions.id = rosters.position_id
JOIN teams ON teams.id = rosters.team_id
LEFT JOIN starters ON starters.roster_id = rosters.id
WHERE rosters.id = $1;

-- name: ListRosterCareerByPlayerID :many
-- 同じ選手の全シーズンのロスター（在籍履歴）。同一シーズンに複数チームの roster がある場合は team_id 順。
SELECT
    rosters.id,
    rosters.season,
    rosters.number,
    rosters.rating,
    rosters.experience,
    positions.name AS position_name,
    sqlc.embed(teams),
    (starters.id IS NOT NULL)::bool AS is_starter
FROM rosters
JOIN positions ON positions.id = rosters.position_id
JOIN teams ON teams.id = rosters.team_id
LEFT JOIN starters ON starters.roster_id = rosters.id
WHERE rosters.player_id = $1
ORDER BY rosters.season DESC, rosters.team_id;
