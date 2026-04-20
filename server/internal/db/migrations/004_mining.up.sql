-- Mining mini-game tables

-- Mining sessions table
CREATE TABLE IF NOT EXISTS mining_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    player_hp INTEGER NOT NULL DEFAULT 100,
    player_max_hp INTEGER NOT NULL DEFAULT 100,
    player_bombs INTEGER NOT NULL DEFAULT 1,
    money_collected REAL NOT NULL DEFAULT 0,
    maze JSONB NOT NULL DEFAULT '{}',
    display_maze JSONB NOT NULL DEFAULT '{}',
    player_x INTEGER NOT NULL DEFAULT 1,
    player_y INTEGER NOT NULL DEFAULT 1,
    exit_x INTEGER NOT NULL DEFAULT 11,
    exit_y INTEGER NOT NULL DEFAULT 11,
    base_level INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    last_move_time TIMESTAMP WITH TIME ZONE
);

-- Mining entities (monsters, items) table
CREATE TABLE IF NOT EXISTS mining_entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES mining_sessions(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('rat','bat','alien','heart','bomb','money')),
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    hp INTEGER NOT NULL DEFAULT 0,
    damage INTEGER NOT NULL DEFAULT 0,
    reward REAL NOT NULL DEFAULT 0,
    alive BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_mining_sessions_player_id ON mining_sessions(player_id);
CREATE INDEX IF NOT EXISTS idx_mining_sessions_planet_id ON mining_sessions(planet_id);
CREATE INDEX IF NOT EXISTS idx_mining_sessions_status ON mining_sessions(status);
CREATE INDEX IF NOT EXISTS idx_mining_entities_session_id ON mining_entities(session_id);

-- Apply updated_at triggers
CREATE TRIGGER update_mining_sessions_updated_at BEFORE UPDATE ON mining_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
