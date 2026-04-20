-- Migration 003 Down: Remove expedition enhancements
DROP INDEX IF EXISTS idx_expeditions_planet_id;
DROP INDEX IF EXISTS idx_expeditions_status;
DROP INDEX IF EXISTS idx_expeditions_discovered_npc_id;
DROP INDEX IF EXISTS idx_npc_planets_owner;

ALTER TABLE expeditions DROP COLUMN IF EXISTS expedition_type;
ALTER TABLE expeditions DROP COLUMN IF EXISTS duration;
ALTER TABLE expeditions DROP COLUMN IF EXISTS elapsed_time;
ALTER TABLE expeditions DROP COLUMN IF EXISTS discovered_npc_id;
ALTER TABLE expeditions DROP COLUMN IF EXISTS actions;
ALTER TABLE expeditions DROP COLUMN IF EXISTS fleet_state;

-- Keep the status column as it was already there
-- ALTER TABLE expeditions DROP COLUMN IF EXISTS status;

ALTER TABLE npc_planets DROP COLUMN IF EXISTS discovered_at;
ALTER TABLE npc_planets DROP COLUMN IF EXISTS poi_type;
