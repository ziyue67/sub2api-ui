-- DB 侧在途预留表：Redis 在途预留不可用时的回落实现（参考 new-api 的 Redis→DB 条件更新回落）。
-- 语义与 billing_cache 的 Lua 预留言一致：按 scope 聚合，scope 内总额超过预算即拒；
-- 行按 (scope, request_id) 唯一，逗期行由 TTL 自愈（每次预留/续期顺手清理）。
-- 幂等：IF NOT EXISTS，可重复执行。
CREATE TABLE IF NOT EXISTS billing_reservations (
    scope      TEXT             NOT NULL,
    request_id TEXT             NOT NULL,
    amount     DOUBLE PRECISION NOT NULL CHECK (amount >= 0),
    expires_at TIMESTAMPTZ      NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    PRIMARY KEY (scope, request_id)
);

CREATE INDEX IF NOT EXISTS idx_billing_reservations_expires_at
    ON billing_reservations (expires_at);
