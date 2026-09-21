-- シードコマンド（cmd/seed）専用のクエリ。

-- name: TruncateSeedTables :exec
TRUNCATE TABLE settings, pf_relations, tf_relations, starters, rosters, origin_rosters,
               users, teams, positions, formations, foul_rules
RESTART IDENTITY CASCADE;

-- name: TeamsExist :one
SELECT EXISTS (SELECT 1 FROM teams);

-- name: PositionsExist :one
SELECT EXISTS (SELECT 1 FROM positions);

-- name: FormationsExist :one
SELECT EXISTS (SELECT 1 FROM formations);

-- name: PfRelationsExist :one
SELECT EXISTS (SELECT 1 FROM pf_relations);

-- name: UsersExist :one
SELECT EXISTS (SELECT 1 FROM users);

-- name: SettingsExist :one
SELECT EXISTS (SELECT 1 FROM settings);

-- name: FoulRulesExist :one
SELECT EXISTS (SELECT 1 FROM foul_rules);

-- name: InsertTeams :copyfrom
INSERT INTO teams (city, name, conference, area, image_file, back_image_file)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertPositions :copyfrom
INSERT INTO positions (abstract_category, category, name, odflg)
VALUES ($1, $2, $3, $4);

-- name: InsertFormations :copyfrom
INSERT INTO formations (odflg, name)
VALUES ($1, $2);

-- name: ListFormations :many
SELECT * FROM formations
ORDER BY id;

-- name: InsertPfRelations :copyfrom
INSERT INTO pf_relations (formation_id, position_category, abstract_category)
VALUES ($1, $2, $3);

-- name: InsertSettings :copyfrom
INSERT INTO settings (user_id, setting_id, flg)
VALUES ($1, $2, $3);

-- name: InsertFoulRules :copyfrom
INSERT INTO foul_rules (english_name, japanese_name, description, status_type, yard_info,
                        lossofdown_flg, af_flg, clock_flg, exit_flg)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
