-- duel_transactions: on-chain prediction/refund/reward events
CREATE TABLE IF NOT EXISTS duel_transactions
(
    signature  String,
    tx_type    UInt8,
    user_id    UUID,
    duel_id    UUID,
    amount     Float64 DEFAULT 0,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (created_at, user_id, duel_id);

-- user_activities: aggregated activity events for retention/engagement metrics
CREATE TABLE IF NOT EXISTS user_activities
(
    user_id        UUID,
    activity_type  String,
    activity_count UInt32,
    created_at     DateTime,
    metadata       String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (created_at, user_id);

-- user_registrations: one row per registered user for cohort analysis
CREATE TABLE IF NOT EXISTS user_registrations
(
    user_id       UUID,
    registered_at DateTime
) ENGINE = ReplacingMergeTree(registered_at)
ORDER BY user_id;

-- token_prices: time-series token price feed
CREATE TABLE IF NOT EXISTS token_prices
(
    token           String,
    symbol          String,
    ts              DateTime,
    usd_price       Float64,
    block_id        Int64,
    decimals        UInt8,
    price_change_24h Float64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (token, ts);

-- admin_audit_logs: moderator action audit trail
CREATE TABLE IF NOT EXISTS admin_audit_logs
(
    moderator_id UUID,
    role         UInt8,
    action       String,
    ip           String,
    metadata     String,
    created_at   DateTime
) ENGINE = MergeTree()
ORDER BY (created_at, moderator_id);
