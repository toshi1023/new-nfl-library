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
ORDER BY players.lastname, players.firstname;
