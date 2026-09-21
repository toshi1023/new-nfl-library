CREATE TABLE origin_rosters (
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    season        INTEGER      NOT NULL,
    team_id       INTEGER      NOT NULL REFERENCES teams (id),
    firstname     VARCHAR(255),
    lastname      VARCHAR(255),
    birthday      VARCHAR(255),
    height        NUMERIC(5,1),
    weight        NUMERIC(5,1),
    college       VARCHAR(255),
    drafted_team  VARCHAR(255),
    drafted_round VARCHAR(255),
    drafted_rank  VARCHAR(255),
    drafted_year  INTEGER,
    position_name VARCHAR(255) NOT NULL,
    number        INTEGER,
    experience    INTEGER,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX origin_rosters_season_team_id_idx ON origin_rosters (season, team_id);

COMMENT ON COLUMN origin_rosters.id            IS 'ID';
COMMENT ON COLUMN origin_rosters.season        IS 'シーズン';
COMMENT ON COLUMN origin_rosters.team_id       IS 'teamsのID';
COMMENT ON COLUMN origin_rosters.firstname     IS '名前';
COMMENT ON COLUMN origin_rosters.lastname      IS '苗字';
COMMENT ON COLUMN origin_rosters.birthday      IS '生年月日';
COMMENT ON COLUMN origin_rosters.height        IS '身長';
COMMENT ON COLUMN origin_rosters.weight        IS '体重';
COMMENT ON COLUMN origin_rosters.college       IS '大学';
COMMENT ON COLUMN origin_rosters.drafted_team  IS 'ドラフトチーム';
COMMENT ON COLUMN origin_rosters.drafted_round IS 'ドラフト巡回数';
COMMENT ON COLUMN origin_rosters.drafted_rank  IS 'ドラフト全体順位';
COMMENT ON COLUMN origin_rosters.drafted_year  IS 'ドラフト年';
COMMENT ON COLUMN origin_rosters.position_name IS 'ポジション名';
COMMENT ON COLUMN origin_rosters.number        IS '背番号';
COMMENT ON COLUMN origin_rosters.experience    IS 'プロ経験年数';
