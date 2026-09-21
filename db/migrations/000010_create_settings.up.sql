CREATE TABLE settings (
    user_id    INTEGER     NOT NULL REFERENCES users (id),
    setting_id INTEGER     NOT NULL CHECK (setting_id >= 0),
    flg        BOOLEAN     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, setting_id)
);

COMMENT ON COLUMN settings.user_id    IS 'usersのID';
COMMENT ON COLUMN settings.setting_id IS '設定の識別子 (1: シーズン選択を表示しない, 2: チーム選択を表示しない ...)';
COMMENT ON COLUMN settings.flg        IS '設定のフラグ';
