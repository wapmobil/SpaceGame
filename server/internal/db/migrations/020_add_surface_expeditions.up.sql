ALTER TABLE planets ADD COLUMN resource_type TEXT NOT NULL DEFAULT 'composite';
ALTER TABLE planets ADD COLUMN max_locations INTEGER NOT NULL DEFAULT 1;

ALTER TABLE planets ADD COLUMN surface_expeditions JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN locations JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN location_buildings JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN expedition_history JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN range_stats JSONB NOT NULL DEFAULT '{}';

CREATE TABLE IF NOT EXISTS surface_expeditions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    progress REAL NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 1800,
    elapsed_time REAL NOT NULL DEFAULT 0,
    discovered_location_id UUID REFERENCES surface_locations(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS surface_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    building_type TEXT,
    building_level INTEGER NOT NULL DEFAULT 1,
    building_active BOOLEAN NOT NULL DEFAULT false,
    source_resource TEXT NOT NULL,
    source_amount REAL NOT NULL,
    source_remaining REAL NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS location_buildings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES surface_locations(id) ON DELETE CASCADE,
    building_type TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT false,
    build_progress REAL NOT NULL DEFAULT 0,
    build_time REAL NOT NULL DEFAULT 0,
    cost_food REAL NOT NULL DEFAULT 0,
    cost_iron REAL NOT NULL DEFAULT 0,
    cost_money REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expedition_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    expedition_type TEXT NOT NULL,
    status TEXT NOT NULL,
    result TEXT NOT NULL,
    discovered TEXT,
    resources_gained JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'expeditions' AND column_name = 'expedition_type'
    ) THEN
        UPDATE expeditions SET expedition_type = 'space_exploration' WHERE expedition_type = 'exploration';
        UPDATE expeditions SET expedition_type = 'space_trade' WHERE expedition_type = 'trade';
        UPDATE expeditions SET expedition_type = 'space_support' WHERE expedition_type = 'support';
    ELSE
        ALTER TABLE expeditions ADD COLUMN expedition_type TEXT NOT NULL DEFAULT 'space_exploration';
    END IF;
END $$;
