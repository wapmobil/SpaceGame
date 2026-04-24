-- Migration 017: Add drill cooldown tracking to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS drill_last_completed TIMESTAMP;
