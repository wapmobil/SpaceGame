-- Migration 005: Ratings, Events, and Statistics

-- Ratings table: precomputed leaderboard data (drop old schema from 001)
DROP TABLE IF EXISTS ratings CASCADE;
CREATE TABLE ratings (
    id SERIAL PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    player_name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('money', 'food', 'ships', 'buildings', 'total_resources')),
    value DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ratings_category ON ratings(category);
CREATE INDEX IF NOT EXISTS idx_ratings_value ON ratings(category, value DESC);
CREATE INDEX IF NOT EXISTS idx_ratings_planet ON ratings(planet_id);

-- Player stats table: persistent statistics
CREATE TABLE IF NOT EXISTS player_stats (
    id SERIAL PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    stat_key TEXT NOT NULL,
    stat_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_stats_player ON player_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_stats_planet ON player_stats(planet_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_stats_key ON player_stats(player_id, stat_key);

-- Daily stats table: daily snapshots of statistics
CREATE TABLE IF NOT EXISTS daily_stats (
    id SERIAL PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    stat_key TEXT NOT NULL,
    stat_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE(player_id, date, stat_key)
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_player ON daily_stats(player_id, date);
CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date);
CREATE INDEX IF NOT EXISTS idx_daily_stats_key ON daily_stats(stat_key);

-- Events table: event log
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_events_player ON events(player_id);
CREATE INDEX IF NOT EXISTS idx_events_planet ON events(planet_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_resolved ON events(resolved);
