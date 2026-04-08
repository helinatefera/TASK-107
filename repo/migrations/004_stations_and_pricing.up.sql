CREATE TABLE stations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id),
    name       TEXT NOT NULL,
    location   TEXT,
    timezone   TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    station_id  UUID NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    device_code TEXT NOT NULL UNIQUE,
    device_type TEXT,
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','maintenance')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE price_templates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id),
    name       TEXT NOT NULL,
    station_id UUID REFERENCES stations(id),
    device_id  UUID REFERENCES devices(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (station_id IS NOT NULL AND device_id IS NULL) OR
        (station_id IS NULL AND device_id IS NOT NULL)
    )
);

CREATE TABLE price_template_versions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id            UUID NOT NULL REFERENCES price_templates(id),
    version_number         INTEGER NOT NULL,
    energy_rate            DECIMAL(12,4) NOT NULL,
    duration_rate          DECIMAL(12,4) NOT NULL,
    service_fee            DECIMAL(12,4) NOT NULL DEFAULT 0,
    apply_tax              BOOLEAN NOT NULL DEFAULT FALSE,
    tax_rate               DECIMAL(8,6),
    status                 TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','inactive')),
    effective_at           TIMESTAMPTZ,
    cloned_from_version_id UUID REFERENCES price_template_versions(id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(template_id, version_number)
);
CREATE INDEX idx_ptv_active ON price_template_versions(template_id, status, effective_at DESC) WHERE status = 'active';

CREATE TABLE tou_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id    UUID NOT NULL REFERENCES price_template_versions(id) ON DELETE CASCADE,
    day_type      TEXT NOT NULL CHECK (day_type IN ('weekday','weekend','holiday')),
    start_time    TIME NOT NULL,
    end_time      TIME NOT NULL,
    energy_rate   DECIMAL(12,4) NOT NULL,
    duration_rate DECIMAL(12,4) NOT NULL,
    CHECK (start_time < end_time)
);
CREATE INDEX idx_tou_version ON tou_rules(version_id, day_type, start_time);

CREATE TABLE order_snapshots (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     TEXT NOT NULL,
    user_id      UUID NOT NULL REFERENCES users(id),
    device_id    UUID NOT NULL REFERENCES devices(id),
    station_id   UUID NOT NULL REFERENCES stations(id),
    version_id   UUID NOT NULL REFERENCES price_template_versions(id),
    energy_rate  DECIMAL(12,4) NOT NULL,
    duration_rate DECIMAL(12,4) NOT NULL,
    service_fee  DECIMAL(12,4) NOT NULL,
    tax_rate     DECIMAL(8,6),
    tou_applied  JSONB,
    energy_kwh   DECIMAL(12,4) NOT NULL,
    duration_min INTEGER NOT NULL,
    subtotal     DECIMAL(12,4) NOT NULL,
    tax_amount   DECIMAL(12,4) NOT NULL DEFAULT 0,
    total        DECIMAL(12,4) NOT NULL,
    order_start  TIMESTAMPTZ NOT NULL,
    order_end    TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user ON order_snapshots(user_id, created_at DESC);
CREATE INDEX idx_orders_device ON order_snapshots(device_id, order_start);
