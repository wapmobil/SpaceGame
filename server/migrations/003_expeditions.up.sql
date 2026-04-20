-- Migration 003: Expeditions Enhancement
-- Adds new columns to expeditions table and enhances npc_planets

-- Enhance expeditions table with new fields
ALTER TABLE expeditions
ADD COLUMN IF NOT EXISTS expedition_type TEXT NOT NULL DEFAULT 'exploration' CHECK (expedition_type IN ('exploration','trade','support')),
ADD COLUMN IF NOT EXISTS duration REAL NOT NULL DEFAULT 3600,
ADD COLUMN IF NOT EXISTS elapsed_time REAL NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS discovered_npc_id UUID REFERENCES npc_planets(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS actions JSONB NOT NULL DEFAULT '[]',
ADD COLUMN IF NOT EXISTS fleet_state JSONB NOT NULL DEFAULT '{}',
ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','at_point','completed','failed','returning'));

-- Add discovered_at to npc_planets
ALTER TABLE npc_planets
ADD COLUMN IF NOT EXISTS discovered_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS poi_type TEXT NOT NULL DEFAULT 'unknown' CHECK (poi_type IN ('abandoned_station','debris','cosmic_debris','asteroids','unknown_planet','alien_base'));

-- Add unique constraint on npc_planets to prevent duplicates
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_npc_planets_id_owner'
    ) THEN
        ALTER TABLE npc_planets ADD CONSTRAINT uq_npc_planets_id_owner UNIQUE (id);
    END IF;
END $$;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_expeditions_planet_id ON expeditions(planet_id);
CREATE INDEX IF NOT EXISTS idx_expeditions_status ON expeditions(status);
CREATE INDEX IF NOT EXISTS idx_expeditions_discovered_npc_id ON expeditions(discovered_npc_id);
CREATE INDEX IF NOT EXISTS idx_npc_planets_owner ON npc_planets(discovered_by);
