CREATE TABLE starters (
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    season     INTEGER     NOT NULL,
    odflg      BOOLEAN     NOT NULL,
    roster_id  INTEGER     NOT NULL REFERENCES rosters (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT starters_roster_id_key UNIQUE (roster_id)
);

CREATE INDEX starters_season_roster_id_idx ON starters (season, roster_id);

COMMENT ON COLUMN starters.id        IS 'ID';
COMMENT ON COLUMN starters.season    IS 'シーズン';
COMMENT ON COLUMN starters.odflg     IS '攻守フラグ (true: Offence, false: Defence)';
COMMENT ON COLUMN starters.roster_id IS 'rostersのID';
