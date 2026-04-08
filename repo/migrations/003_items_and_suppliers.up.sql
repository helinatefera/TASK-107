CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    parent_id  UUID REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE units_of_measure (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    symbol     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE unit_conversions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_unit_id UUID NOT NULL REFERENCES units_of_measure(id),
    to_unit_id   UUID NOT NULL REFERENCES units_of_measure(id),
    factor       DECIMAL(18,6) NOT NULL CHECK (factor > 0),
    UNIQUE(from_unit_id, to_unit_id)
);

CREATE TABLE items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku          TEXT NOT NULL UNIQUE,
    item_name    TEXT NOT NULL,
    category_id  UUID REFERENCES categories(id),
    base_unit_id UUID NOT NULL REFERENCES units_of_measure(id),
    description  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    tax_id          TEXT,
    contact_email   TEXT,
    address         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_suppliers_dedup ON suppliers(normalized_name, tax_id) WHERE tax_id IS NOT NULL;

CREATE TABLE carriers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    tax_id          TEXT,
    contact_email   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_carriers_dedup ON carriers(normalized_name, tax_id) WHERE tax_id IS NOT NULL;
