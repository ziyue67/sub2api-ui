-- Durable admission reservations used only while Redis is unavailable.
-- Redis remains the hot-path reservation ledger; this table prevents the
-- fail-open window where concurrent requests could all spend the same wallet.

CREATE TABLE IF NOT EXISTS billing_reservation_fallback_state (
    singleton      BOOLEAN PRIMARY KEY DEFAULT TRUE,
    fallback_until TIMESTAMPTZ NOT NULL DEFAULT '-infinity',
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_billing_reservation_fallback_singleton CHECK (singleton)
);

INSERT INTO billing_reservation_fallback_state (singleton)
VALUES (TRUE)
ON CONFLICT (singleton) DO NOTHING;

CREATE TABLE IF NOT EXISTS billing_reservations (
    scope      TEXT NOT NULL,
    request_id VARCHAR(128) NOT NULL,
    amount     NUMERIC(20, 8) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (scope, request_id),
    CONSTRAINT chk_billing_reservations_amount_positive CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_billing_reservations_scope_expiry
    ON billing_reservations(scope, expires_at);

CREATE INDEX IF NOT EXISTS idx_billing_reservations_expiry
    ON billing_reservations(expires_at);
