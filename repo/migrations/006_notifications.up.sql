CREATE TABLE notification_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT NOT NULL UNIQUE,
    title_tmpl      TEXT NOT NULL,
    body_tmpl       TEXT NOT NULL,
    default_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notification_subscriptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES notification_templates(id),
    opted_out   BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, template_id)
);

CREATE TABLE notification_jobs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    template_id     UUID NOT NULL REFERENCES notification_templates(id),
    params          JSONB,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','delivered','suppressed','failed')),
    suppress_reason TEXT,
    scheduled_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notif_jobs_pending ON notification_jobs(status, scheduled_at) WHERE status = 'pending';

CREATE TABLE messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES notification_templates(id),
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    read        BOOLEAN NOT NULL DEFAULT FALSE,
    dismissed   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_user_inbox ON messages(user_id, created_at DESC) WHERE dismissed = FALSE;

CREATE TABLE delivery_stats (
    id          BIGSERIAL PRIMARY KEY,
    template_id UUID NOT NULL REFERENCES notification_templates(id),
    date        DATE NOT NULL,
    generated   INTEGER NOT NULL DEFAULT 0,
    delivered   INTEGER NOT NULL DEFAULT 0,
    opened      INTEGER NOT NULL DEFAULT 0,
    dismissed   INTEGER NOT NULL DEFAULT 0,
    UNIQUE(template_id, date)
);
