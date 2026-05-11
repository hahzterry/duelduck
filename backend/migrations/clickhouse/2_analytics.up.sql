-- Add project_id to duel_transactions for per-project analytics queries
ALTER TABLE duel_transactions ADD COLUMN IF NOT EXISTS project_id UUID DEFAULT '00000000-0000-0000-0000-000000000000';
