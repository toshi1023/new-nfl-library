CREATE TABLE players (
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firstname     VARCHAR(255) NOT NULL,
    lastname      VARCHAR(255) NOT NULL,
    birthday      VARCHAR(255) NOT NULL,
    height        NUMERIC(5,1),
    weight        NUMERIC(5,1),
    college       VARCHAR(255),
    drafted_team  VARCHAR(255),
    drafted_round VARCHAR(255),
    drafted_rank  VARCHAR(255),
    drafted_year  INTEGER,
    image_file    VARCHAR(255),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON COLUMN players.id            IS 'ID';
COMMENT ON COLUMN players.firstname     IS '名前';
COMMENT ON COLUMN players.lastname      IS '苗字';
COMMENT ON COLUMN players.birthday      IS '生年月日';
COMMENT ON COLUMN players.height        IS '身長';
COMMENT ON COLUMN players.weight        IS '体重';
COMMENT ON COLUMN players.college       IS '大学';
COMMENT ON COLUMN players.drafted_team  IS 'ドラフトチーム';
COMMENT ON COLUMN players.drafted_round IS 'ドラフト巡回数';
COMMENT ON COLUMN players.drafted_rank  IS 'ドラフト全体順位';
COMMENT ON COLUMN players.drafted_year  IS 'ドラフト年';
COMMENT ON COLUMN players.image_file    IS '画像名';
