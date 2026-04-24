-- Migration 016: Add drill mini-game tables
CREATE TABLE IF NOT EXISTS drill_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    drill_hp INTEGER NOT NULL DEFAULT 100,
    drill_max_hp INTEGER NOT NULL DEFAULT 100,
    depth INTEGER NOT NULL DEFAULT 0,
    drill_x INTEGER NOT NULL DEFAULT 10,
    world_width INTEGER NOT NULL DEFAULT 20,
    speed_multiplier REAL NOT NULL DEFAULT 1.0,
    resources JSONB NOT NULL DEFAULT '[]',
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'failed')),
    total_earned REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    last_move_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS drill_world (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES drill_sessions(id) ON DELETE CASCADE,
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    cell_type TEXT NOT NULL,
    resource_type TEXT,
    resource_amount REAL NOT NULL DEFAULT 0,
    resource_value REAL NOT NULL DEFAULT 0,
    extracted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_drill_sessions_player_id ON drill_sessions(player_id);
CREATE INDEX IF NOT EXISTS idx_drill_sessions_planet_id ON drill_sessions(planet_id);
CREATE INDEX IF NOT EXISTS idx_drill_sessions_status ON drill_sessions(status);
CREATE INDEX IF NOT EXISTS idx_drill_world_session_id ON drill_world(session_id);
