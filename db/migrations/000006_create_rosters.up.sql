CREATE TABLE rosters (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    season      INTEGER     NOT NULL,
    team_id     INTEGER     NOT NULL REFERENCES teams (id),
    player_id   INTEGER     NOT NULL REFERENCES players (id),
    position_id INTEGER     NOT NULL REFERENCES positions (id),
    number      INTEGER,
    rating      INTEGER,
    experience  INTEGER,
    madden_attributes JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rosters_season_team_player_key UNIQUE (season, team_id, player_id)
);

CREATE INDEX rosters_season_team_id_idx ON rosters (season, team_id);
CREATE INDEX rosters_player_id_idx      ON rosters (player_id);
CREATE INDEX rosters_position_id_idx    ON rosters (position_id);

COMMENT ON COLUMN rosters.id          IS 'ID';
COMMENT ON COLUMN rosters.season      IS 'シーズン';
COMMENT ON COLUMN rosters.team_id     IS 'teamsのID';
COMMENT ON COLUMN rosters.player_id   IS 'playersのID';
COMMENT ON COLUMN rosters.position_id IS 'positionsのID';
COMMENT ON COLUMN rosters.number      IS '背番号';
COMMENT ON COLUMN rosters.rating      IS '能力値';
COMMENT ON COLUMN rosters.experience  IS 'プロ経験年数';
COMMENT ON COLUMN rosters.madden_attributes IS 'Madden NFL の評価値CSVのうち、他の列に対応しない項目一式（JSON）';
