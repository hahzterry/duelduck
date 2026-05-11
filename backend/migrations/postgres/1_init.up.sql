-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- users: prediction-api user model (simpler than duel-duck-api)
CREATE TABLE IF NOT EXISTS users (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id     UUID,
    email          VARCHAR(255)  UNIQUE,
    role           INTEGER       NOT NULL DEFAULT 0,
    wallet_address VARCHAR(255),
    is_active      BOOLEAN       NOT NULL DEFAULT true,
    created_at     TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- projects: partner projects that host prediction rooms
CREATE TABLE IF NOT EXISTS projects (
    id                       UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id               UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key                  VARCHAR(255)  NOT NULL UNIQUE,
    is_users_duels_enabled   BOOLEAN       NOT NULL DEFAULT false,
    is_self_resolved_enabled BOOLEAN       NOT NULL DEFAULT false,
    is_blocked               BOOLEAN       NOT NULL DEFAULT false,
    created_at               TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- duels: prediction rooms
CREATE TABLE IF NOT EXISTS duels (
    id                    UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id              UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id            UUID           NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    resolved_by           UUID           REFERENCES users(id),
    approved_by           UUID           REFERENCES users(id),
    resolved_at           TIMESTAMP WITHOUT TIME ZONE,
    is_owner_resolving    BOOLEAN        NOT NULL DEFAULT false,
    is_self_resolved      BOOLEAN        NOT NULL DEFAULT false,
    room_number           BIGINT         UNIQUE,
    room_token_pda        TEXT,
    symbol                TEXT           NOT NULL,
    players_count         INTEGER        NOT NULL DEFAULT 0,
    refunded_players_count INTEGER       NOT NULL DEFAULT 0,
    winners_count         INTEGER        NOT NULL DEFAULT 0,
    username              VARCHAR(17)    NOT NULL,
    status                INTEGER        NOT NULL DEFAULT 0,
    logo_url              TEXT,
    question              TEXT,
    slug                  TEXT,
    source_of_truth       TEXT,
    duel_price            DECIMAL(15,9)  NOT NULL,
    usd_price             DECIMAL(15,9),
    commission            INTEGER        NOT NULL DEFAULT 0,
    commission_rate       INTEGER        NOT NULL DEFAULT 0,
    duel_info             JSONB,
    final_result          INTEGER,
    cancellation_reason   TEXT,
    category_id           UUID,
    category_name         TEXT,
    tournament_id         UUID,
    deadline              TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at            TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- players: participants in a duel
CREATE TABLE IF NOT EXISTS players (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    duel_id      UUID          NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    answer       INTEGER       NOT NULL DEFAULT 0,
    is_winner    BOOLEAN       NOT NULL DEFAULT false,
    win_amount   DECIMAL(15,9) NOT NULL DEFAULT 0,
    final_status SMALLINT      NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- transactions: on-chain signatures registry
CREATE TABLE IF NOT EXISTS transactions (
    signature CHAR(88)  PRIMARY KEY,
    tx_type   SMALLINT  NOT NULL CHECK (tx_type IN (1, 2, 3, 4))
);

-- coins: CoinMarketCap coin list
CREATE TABLE IF NOT EXISTS coins (
    id        BIGINT        PRIMARY KEY,
    rank      BIGINT        NOT NULL,
    name      VARCHAR(100)  NOT NULL,
    symbol    VARCHAR(50)   NOT NULL,
    slug      VARCHAR(100)  NOT NULL,
    image_url VARCHAR(200)  NOT NULL
);

-- solana_tokens: SPL token registry with price/metadata
CREATE TABLE IF NOT EXISTS solana_tokens (
    mint       VARCHAR(44)   PRIMARY KEY,
    name       VARCHAR(128)  NOT NULL,
    symbol     VARCHAR(10)   NOT NULL,
    decimals   SMALLINT      NOT NULL,
    image_url  TEXT,
    market_cap DOUBLE PRECISION NOT NULL DEFAULT 0,
    usd_price  DECIMAL(15,9) NOT NULL DEFAULT 0,
    is_verified BOOLEAN      NOT NULL DEFAULT false,
    program_id VARCHAR(100)  NOT NULL DEFAULT 'empty',
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS solana_tokens_verified_symbol_trgm_idx
    ON solana_tokens USING GIN (symbol gin_trgm_ops) WHERE is_verified = true;
CREATE INDEX IF NOT EXISTS solana_tokens_verified_name_trgm_idx
    ON solana_tokens USING GIN (name gin_trgm_ops) WHERE is_verified = true;

-- duel_status_history: audit trail of duel status transitions
CREATE TABLE IF NOT EXISTS duel_status_history (
    id         BIGSERIAL PRIMARY KEY,
    duel_id    UUID      NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    status     INTEGER   NOT NULL,
    changed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_duel_status_history_changed_at
    ON duel_status_history (changed_at);
CREATE INDEX IF NOT EXISTS idx_duel_status_history_duel_id_changed_at
    ON duel_status_history (duel_id, changed_at);

-- project_commission_claims: partner and DD profit payout records
CREATE TABLE IF NOT EXISTS project_commission_claims (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    type           VARCHAR(20)   NOT NULL,
    project_id     UUID          REFERENCES projects(id),
    period_start   DATE          NOT NULL,
    period_end     DATE          NOT NULL,
    gross_usd      DECIMAL(15,6) NOT NULL,
    platform_rate  DECIMAL(5,4)  NOT NULL,
    amount_usd     DECIMAL(15,6) NOT NULL,
    wallet_address TEXT          NOT NULL,
    tx_hashes      JSONB         NOT NULL DEFAULT '{}',
    claimed_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- project_commission_accruals: gross commission earned per duel
CREATE TABLE IF NOT EXISTS project_commission_accruals (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id     UUID          NOT NULL,
    duel_id        UUID          NOT NULL UNIQUE,
    symbol         VARCHAR(20)   NOT NULL,
    commission_raw BIGINT        NOT NULL,
    commission_usd DECIMAL(15,6) NOT NULL,
    billing_month  DATE          NOT NULL,
    accrued_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pca_project_billing_month
    ON project_commission_accruals (project_id, billing_month);
