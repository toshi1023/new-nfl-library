CREATE TABLE positions (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    abstract_category INTEGER      NOT NULL,
    category          INTEGER      NOT NULL,
    name              VARCHAR(255) NOT NULL,
    odflg             BOOLEAN      NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON COLUMN positions.id                IS 'ID';
COMMENT ON COLUMN positions.abstract_category IS '抽象カテゴリー';
COMMENT ON COLUMN positions.category          IS 'カテゴリー';
COMMENT ON COLUMN positions.name              IS '名前';
COMMENT ON COLUMN positions.odflg             IS '攻守フラグ (true: Offence, false: Defence)';
