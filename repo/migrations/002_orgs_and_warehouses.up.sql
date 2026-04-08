CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  UUID REFERENCES organizations(id),
    org_code   TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    tax_id     TEXT,
    address    TEXT,
    timezone   TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orgs_parent ON organizations(parent_id);

ALTER TABLE users ADD CONSTRAINT fk_users_org FOREIGN KEY (org_id) REFERENCES organizations(id);

CREATE TABLE warehouses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id),
    name       TEXT NOT NULL,
    address    TEXT,
    timezone   TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE zones (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    zone_type    TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE bins (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id      UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    bin_code     TEXT NOT NULL,
    capacity     INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(warehouse_id, bin_code)
);
