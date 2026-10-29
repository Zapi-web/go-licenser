-- +goose Up
CREATE TABLE IF NOT EXISTS licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_key TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    hwid TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX licenses_expires_at_idx ON licenses (expires_at)
    WHERE expires_at IS NOT NULL AND status = 'active';

-- +goose Down
DROP TABLE IF EXISTS licenses;
