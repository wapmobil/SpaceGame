-- Rollback: drop new tables, restore old columns
DROP TABLE IF EXISTS expedition_events;
DROP TABLE IF EXISTS expedition_chains;
ALTER TABLE planets ADD COLUMN IF NOT EXISTS surface_expeditions JSONB DEFAULT '[]';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS expedition_history JSONB DEFAULT '[]';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS range_stats JSONB DEFAULT '{}';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS max_expeditions INTEGER DEFAULT 1;
