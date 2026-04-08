CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   TEXT NOT NULL,
    old_value   JSONB,
    new_value   JSONB,
    ip_address  INET,
    request_id  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);

CREATE TABLE app_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed global configurations managed by administrators
INSERT INTO app_config (key, value) VALUES
    ('tax_rate', '0.08'),
    ('system.name', 'ChargeOps'),
    ('system.version', '1.0.0');

CREATE TABLE request_metrics (
    id          BIGSERIAL PRIMARY KEY,
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms  INTEGER NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_metrics_time ON request_metrics(recorded_at DESC);
