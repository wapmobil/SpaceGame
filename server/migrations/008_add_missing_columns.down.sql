ALTER TABLE planets DROP COLUMN IF EXISTS fleet_state;
ALTER TABLE planets DROP COLUMN IF EXISTS shipyard_state;
ALTER TABLE research DROP COLUMN IF EXISTS total_time;
ALTER TABLE research DROP COLUMN IF EXISTS start_time;
ALTER TABLE research DROP COLUMN IF EXISTS level;
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS buildings_planet_type;
