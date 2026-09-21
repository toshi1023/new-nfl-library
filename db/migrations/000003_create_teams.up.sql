CREATE TABLE teams (
    id              INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    city            VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    conference      SMALLINT     NOT NULL DEFAULT 1,
    area            SMALLINT     NOT NULL DEFAULT 1,
    image_file      VARCHAR(255),
    back_image_file VARCHAR(255),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON COLUMN teams.id              IS 'ID';
COMMENT ON COLUMN teams.city            IS '拠点名';
COMMENT ON COLUMN teams.name            IS 'チーム名';
COMMENT ON COLUMN teams.conference      IS '所属カンファレンス (1: AFC, 2: NFC)';
COMMENT ON COLUMN teams.area            IS '所属地区 (1: 東, 2: 北, 3: 南, 4: 西)';
COMMENT ON COLUMN teams.image_file      IS '画像名';
COMMENT ON COLUMN teams.back_image_file IS '背景用画像名';

-- users.favorite_team は teams 作成後に外部キーを張る
ALTER TABLE users
    ADD CONSTRAINT users_favorite_team_fkey
    FOREIGN KEY (favorite_team) REFERENCES teams (id);
