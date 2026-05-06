-- Phase 1: Create new expedition chains schema first, then drop old schema
-- New tables are created BEFORE dropping old ones so we don't lose data if migration fails mid-way

CREATE TABLE IF NOT EXISTS expedition_chains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'failed')),
    event_count INTEGER NOT NULL DEFAULT 0,
    current_event_index INTEGER NOT NULL DEFAULT 0,
    discovered_location_id UUID REFERENCES surface_locations(id) ON DELETE SET NULL,
    inventory JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expedition_chains_planet ON expedition_chains(planet_id);
CREATE INDEX idx_expedition_chains_owner ON expedition_chains(owner_id);
CREATE INDEX idx_expedition_chains_status ON expedition_chains(status);

CREATE TABLE IF NOT EXISTS expedition_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id UUID NOT NULL REFERENCES expedition_chains(id) ON DELETE CASCADE,
    event_id TEXT NOT NULL,
    description TEXT NOT NULL,
    choices JSONB NOT NULL DEFAULT '[]',
    immediate_reward JSONB NOT NULL DEFAULT '{}',
    is_end BOOLEAN NOT NULL DEFAULT false,
    location_reward TEXT,
    player_choice INTEGER,
    rewards_received JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expedition_events_chain ON expedition_events(chain_id);

-- Drop old JSONB columns from planets
ALTER TABLE planets DROP COLUMN IF EXISTS surface_expeditions;
ALTER TABLE planets DROP COLUMN IF EXISTS expedition_history;
ALTER TABLE planets DROP COLUMN IF EXISTS range_stats;
ALTER TABLE planets DROP COLUMN IF EXISTS max_expeditions;

-- Drop old tables
DROP TABLE IF EXISTS surface_expeditions;
DROP TABLE IF EXISTS expedition_history;
