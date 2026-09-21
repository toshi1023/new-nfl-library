CREATE TABLE foul_rules (
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    english_name   VARCHAR(255) NOT NULL,
    japanese_name  VARCHAR(255) NOT NULL,
    description    TEXT         NOT NULL,
    status_type    INTEGER      NOT NULL,
    yard_info      INTEGER      NOT NULL,
    lossofdown_flg BOOLEAN      NOT NULL DEFAULT false,
    af_flg         BOOLEAN      NOT NULL DEFAULT false,
    clock_flg      BOOLEAN      NOT NULL DEFAULT false,
    exit_flg       BOOLEAN      NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON COLUMN foul_rules.id             IS 'ID';
COMMENT ON COLUMN foul_rules.english_name   IS '英語表記';
COMMENT ON COLUMN foul_rules.japanese_name  IS '日本語表記';
COMMENT ON COLUMN foul_rules.description    IS '説明';
COMMENT ON COLUMN foul_rules.status_type    IS '攻守ステータス (0: 共通, 1: 攻撃, 2: 守備, 3: キック, 99: 全て)';
COMMENT ON COLUMN foul_rules.yard_info      IS '罰則ヤード';
COMMENT ON COLUMN foul_rules.lossofdown_flg IS 'ダウンの喪失フラグ';
COMMENT ON COLUMN foul_rules.af_flg         IS '自動ファーストダウン更新フラグ';
COMMENT ON COLUMN foul_rules.clock_flg      IS 'ゲームクロックの減算フラグ';
COMMENT ON COLUMN foul_rules.exit_flg       IS '退場フラグ';
