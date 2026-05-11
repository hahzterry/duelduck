-- Each project now has its own custodial Solana wallet for reward distribution.
-- The keypair is generated at project creation time and stored in Vault.
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS wallet_address TEXT NOT NULL;
