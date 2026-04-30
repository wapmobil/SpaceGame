-- Migration 017 down: Remove drill cooldown tracking from players table
ALTER TABLE players DROP COLUMN IF EXISTS drill_last_completed;
