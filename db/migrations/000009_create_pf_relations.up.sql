CREATE TABLE pf_relations (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    formation_id      INTEGER     NOT NULL REFERENCES formations (id),
    position_category INTEGER     NOT NULL,
    abstract_category INTEGER     NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX pf_relations_formation_id_idx ON pf_relations (formation_id);

COMMENT ON COLUMN pf_relations.id                IS 'ID';
COMMENT ON COLUMN pf_relations.formation_id      IS 'formationsのID';
COMMENT ON COLUMN pf_relations.position_category IS 'positionsのcategory';
COMMENT ON COLUMN pf_relations.abstract_category IS 'positionsのabstract_category';
