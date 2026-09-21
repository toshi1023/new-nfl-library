CREATE TABLE users (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name              VARCHAR(255) NOT NULL,
    email             VARCHAR(255) NOT NULL UNIQUE,
    email_verified_at TIMESTAMPTZ,
    password          VARCHAR(255) NOT NULL,
    remember_token    VARCHAR(100),
    favorite_season   INTEGER,
    favorite_team     INTEGER,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON COLUMN users.id                IS 'ID';
COMMENT ON COLUMN users.name              IS 'ユーザ名';
COMMENT ON COLUMN users.email             IS 'メールアドレス';
COMMENT ON COLUMN users.email_verified_at IS 'メールアドレス(確認用)';
COMMENT ON COLUMN users.password          IS 'パスワード';
COMMENT ON COLUMN users.remember_token    IS 'ログイン時間維持用トークン';
COMMENT ON COLUMN users.favorite_season   IS 'お気に入りシーズン';
COMMENT ON COLUMN users.favorite_team     IS 'お気に入りチーム(teamsのid)';
