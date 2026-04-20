CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Planets table
CREATE TABLE IF NOT EXISTS planets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    resources JSONB NOT NULL DEFAULT '{"food":0,"composite":0,"mechanisms":0,"reagents":0,"energy":0,"max_energy":100,"money":0,"alien_tech":0}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Buildings table
CREATE TABLE IF NOT EXISTS buildings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('farm','solar','storage','base','factory','energy_storage','shipyard','comcenter')),
    level INTEGER NOT NULL DEFAULT 1,
    build_progress REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Research table
CREATE TABLE IF NOT EXISTS research (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    tech_id TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    in_progress BOOLEAN NOT NULL DEFAULT FALSE,
    progress REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(planet_id, tech_id)
);

-- Ships table
CREATE TABLE IF NOT EXISTS ships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('trade','small','interceptor','corvette','frigate','cruiser')),
    hp INTEGER NOT NULL DEFAULT 100,
    armor INTEGER NOT NULL DEFAULT 0,
    weapons JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Fleets table
CREATE TABLE IF NOT EXISTS fleets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    ships JSONB NOT NULL DEFAULT '[]',
    energy INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Expeditions table
CREATE TABLE IF NOT EXISTS expeditions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    fleet_id UUID REFERENCES fleets(id) ON DELETE SET NULL,
    target TEXT NOT NULL,
    progress REAL NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','completed','failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Battles table
CREATE TABLE IF NOT EXISTS battles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attacker_planet_id UUID NOT NULL REFERENCES planets(id),
    defender_planet_id UUID NOT NULL REFERENCES planets(id),
    grid JSONB NOT NULL DEFAULT '{}',
    phase TEXT NOT NULL DEFAULT 'setup',
    turn INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Market orders table
CREATE TABLE IF NOT EXISTS market_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    resource TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('buy','sell')),
    amount INTEGER NOT NULL,
    price REAL NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- NPC planets table
CREATE TABLE IF NOT EXISTS npc_planets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    resources JSONB NOT NULL DEFAULT '{"food":0,"composite":0,"mechanisms":0,"reagents":0,"energy":0,"max_energy":100,"money":0,"alien_tech":0}',
    enemy_fleet JSONB NOT NULL DEFAULT '[]',
    discovered_by TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    money_rank INTEGER NOT NULL DEFAULT 0,
    food_rank INTEGER NOT NULL DEFAULT 0,
    ship_rank INTEGER NOT NULL DEFAULT 0,
    building_rank INTEGER NOT NULL DEFAULT 0,
    total_resources_rank INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_planets_updated_at BEFORE UPDATE ON planets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_buildings_updated_at BEFORE UPDATE ON buildings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_research_updated_at BEFORE UPDATE ON research FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ships_updated_at BEFORE UPDATE ON ships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_fleets_updated_at BEFORE UPDATE ON fleets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_expeditions_updated_at BEFORE UPDATE ON expeditions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_market_orders_updated_at BEFORE UPDATE ON market_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_npc_planets_updated_at BEFORE UPDATE ON npc_planets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ratings_updated_at BEFORE UPDATE ON ratings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
