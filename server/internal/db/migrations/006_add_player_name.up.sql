-- Add name column to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS name TEXT DEFAULT '';
