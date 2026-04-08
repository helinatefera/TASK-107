CREATE TABLE carousel_slots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    title       TEXT NOT NULL,
    image_url   TEXT,
    link_url    TEXT,
    priority    INTEGER NOT NULL DEFAULT 0,
    target_role TEXT CHECK (target_role IN ('guest','user','merchant','administrator')),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (start_time < end_time)
);

CREATE TABLE campaign_placements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    name        TEXT NOT NULL,
    content     JSONB NOT NULL,
    priority    INTEGER NOT NULL DEFAULT 0,
    target_role TEXT CHECK (target_role IN ('guest','user','merchant','administrator')),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (start_time < end_time)
);

CREATE TABLE hot_rankings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    entity_type TEXT NOT NULL,
    entity_id   UUID NOT NULL,
    score       INTEGER NOT NULL DEFAULT 0,
    priority    INTEGER NOT NULL DEFAULT 0,
    target_role TEXT CHECK (target_role IN ('guest','user','merchant','administrator')),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (start_time < end_time)
);
