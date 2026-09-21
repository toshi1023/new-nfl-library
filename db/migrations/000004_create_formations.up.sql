CREATE TABLE formations (
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    odflg      BOOLEAN      NOT NULL DEFAULT true,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON COLUMN formations.id    IS 'ID';
COMMENT ON COLUMN formations.odflg IS '攻守フラグ (true: Offence, false: Defence)';
COMMENT ON COLUMN formations.name  IS '名前';
