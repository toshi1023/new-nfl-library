CREATE TABLE tf_relations (
    id           INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    season       INTEGER     NOT NULL,
    team_id      INTEGER     NOT NULL REFERENCES teams (id),
    odflg        BOOLEAN     NOT NULL,
    formation_id INTEGER     NOT NULL REFERENCES formations (id),
    abstract_flg BOOLEAN     NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX tf_relations_season_team_id_idx ON tf_relations (season, team_id);
CREATE INDEX tf_relations_formation_id_idx   ON tf_relations (formation_id);

COMMENT ON COLUMN tf_relations.id           IS 'ID';
COMMENT ON COLUMN tf_relations.season       IS 'シーズン';
COMMENT ON COLUMN tf_relations.team_id      IS 'teamsのID';
COMMENT ON COLUMN tf_relations.odflg        IS '攻守フラグ (true: Offence, false: Defence)';
COMMENT ON COLUMN tf_relations.formation_id IS 'formationsのID';
COMMENT ON COLUMN tf_relations.abstract_flg IS '結合フラグ (pf_relationsとpositionsをabstract_categoryで結合するかどうか)';
