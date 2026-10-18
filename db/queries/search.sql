-- GET /api/v1/search/seasons・GET /api/v1/search/teams（検索UIの選択肢を返すAPI）専用のクエリ。

-- name: ListRosterSeasons :many
SELECT DISTINCT season FROM rosters ORDER BY season DESC;

-- name: ListTeams :many
SELECT id, city, name, conference, area, image_file, back_image_file
FROM teams
ORDER BY id;
