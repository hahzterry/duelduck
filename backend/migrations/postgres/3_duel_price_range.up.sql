-- Add price type to duels: 'fixed' (default) or 'range'
ALTER TABLE duels
    ADD COLUMN IF NOT EXISTS price_type VARCHAR(10) NOT NULL DEFAULT 'fixed',
    ADD COLUMN IF NOT EXISTS min_price   DECIMAL(15,9),
    ADD COLUMN IF NOT EXISTS max_price   DECIMAL(15,9);

-- Store each player's actual chosen price (populated for price_range duels)
ALTER TABLE players
    ADD COLUMN IF NOT EXISTS paid_price DECIMAL(15,9);

-- Backfill players.paid_price from parent duel for existing fixed-type duels
UPDATE players p
SET paid_price = d.duel_price
FROM duels d
WHERE p.duel_id = d.id;
