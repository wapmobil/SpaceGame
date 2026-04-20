-- Migration 005: Ratings, Events, and Statistics

-- Ratings table: precomputed leaderboard data
CREATE TABLE IF NOT EXISTS ratings (
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

-- Functions

-- Compute ratings for a given category
CREATE OR REPLACE FUNCTION compute_ratings_for_category(cat TEXT)
RETURNS VOID AS $$
DECLARE
    rec RECORD;
    planet_rec RECORD;
    total_resource_value DOUBLE PRECISION;
    money_val DOUBLE PRECISION;
    food_val DOUBLE PRECISION;
    ship_count INTEGER;
    building_count INTEGER;
    total_buildings INTEGER;
BEGIN
    -- Delete existing ratings for this category
    DELETE FROM ratings WHERE category = cat;

    FOR rec IN SELECT id, player_id, name FROM planets ORDER BY updated_at DESC
    LOOP
        planet_rec := rec;

        CASE cat
            WHEN 'money' THEN
                money_val := 0;
                SELECT COALESCE(resources->>'money', '0')::DOUBLE PRECISION
                INTO money_val
                FROM planets WHERE id = planet_rec.id;
                INSERT INTO ratings (planet_id, player_name, category, value)
                VALUES (planet_rec.id, planet_rec.name, cat, money_val);

            WHEN 'food' THEN
                food_val := 0;
                SELECT COALESCE(resources->>'food', '0')::DOUBLE PRECISION
                INTO food_val
                FROM planets WHERE id = planet_rec.id;
                INSERT INTO ratings (planet_id, player_name, category, value)
                VALUES (planet_rec.id, planet_rec.name, cat, food_val);

            WHEN 'ships' THEN
                ship_count := 0;
                SELECT COALESCE((fleet_state::JSONB)->'total_ships', '0')::INTEGER
                INTO ship_count
                FROM planets WHERE id = planet_rec.id;
                INSERT INTO ratings (planet_id, player_name, category, value)
                VALUES (planet_rec.id, planet_rec.name, cat, ship_count::DOUBLE PRECISION);

            WHEN 'buildings' THEN
                total_buildings := 0;
                SELECT COALESCE((buildings::JSONB)->'farm', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'solar', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'storage', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'base', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'factory', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'energy_storage', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'shipyard', '0')::INTEGER +
                       COALESCE((buildings::JSONB)->'comcenter', '0')::INTEGER
                INTO total_buildings
                FROM planets WHERE id = planet_rec.id;
                INSERT INTO ratings (planet_id, player_name, category, value)
                VALUES (planet_rec.id, planet_rec.name, cat, total_buildings::DOUBLE PRECISION);

            WHEN 'total_resources' THEN
                SELECT COALESCE(resources->>'food', '0')::DOUBLE PRECISION +
                       COALESCE(resources->>'composite', '0')::DOUBLE PRECISION +
                       COALESCE(resources->>'mechanisms', '0')::DOUBLE PRECISION +
                       COALESCE(resources->>'reagents', '0')::DOUBLE PRECISION +
                       COALESCE(resources->>'money', '0')::DOUBLE PRECISION +
                       COALESCE(resources->>'alien_tech', '0')::DOUBLE PRECISION
                INTO total_resource_value
                FROM planets WHERE id = planet_rec.id;
                INSERT INTO ratings (planet_id, player_name, category, value)
                VALUES (planet_rec.id, planet_rec.name, cat, total_resource_value);
        END CASE;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Get player rank for a category
CREATE OR REPLACE FUNCTION get_player_rank(cat TEXT, p_planet_id UUID)
RETURNS TABLE(rank INTEGER, planet_id UUID, player_name TEXT, value DOUBLE PRECISION) AS $$
BEGIN
    RETURN QUERY
    SELECT
        sub.rank,
        r.planet_id,
        r.player_name,
        r.value
    FROM (
        SELECT planet_id, player_name, value,
               ROW_NUMBER() OVER (ORDER BY value DESC) AS rank
        FROM ratings
        WHERE category = cat
    ) sub
    JOIN ratings r ON r.planet_id = sub.planet_id
    WHERE sub.rank IS NOT NULL
      AND r.category = cat
      AND r.planet_id = p_planet_id;
END;
$$ LANGUAGE plpgsql;

-- Log an event
CREATE OR REPLACE FUNCTION log_event(
    p_planet_id UUID,
    p_player_id UUID,
    p_event_type TEXT,
    p_description TEXT
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO events (planet_id, player_id, event_type, description)
    VALUES (p_planet_id, p_player_id, p_event_type, p_description);
END;
$$ LANGUAGE plpgsql;

-- Get player stats
CREATE OR REPLACE FUNCTION get_player_stats(p_player_id UUID)
RETURNS TABLE(stat_key TEXT, stat_value DOUBLE PRECISION, updated_at TIMESTAMP WITH TIME ZONE) AS $$
BEGIN
    RETURN QUERY
    SELECT stat_key, stat_value, updated_at
    FROM player_stats
    WHERE player_id = p_player_id
    ORDER BY stat_key;
END;
$$ LANGUAGE plpgsql;

-- Get daily stats for a player and date range
CREATE OR REPLACE FUNCTION get_daily_stats(
    p_player_id UUID,
    p_start_date DATE,
    p_end_date DATE
)
RETURNS TABLE(stat_key TEXT, date DATE, stat_value DOUBLE PRECISION) AS $$
BEGIN
    RETURN QUERY
    SELECT stat_key, date, stat_value
    FROM daily_stats
    WHERE player_id = p_player_id
      AND date BETWEEN p_start_date AND p_end_date
    ORDER BY date DESC, stat_key;
END;
$$ LANGUAGE plpgsql;

-- Reset daily stats (called by scheduler at 6:00 AM)
CREATE OR REPLACE FUNCTION reset_daily_stats(p_date DATE)
RETURNS VOID AS $$
DECLARE
    rec RECORD;
    current_value DOUBLE PRECISION;
BEGIN
    -- Copy current player_stats to daily_stats for the given date
    FOR rec IN SELECT player_id, planet_id, stat_key, stat_value FROM player_stats
    LOOP
        INSERT INTO daily_stats (player_id, planet_id, date, stat_key, stat_value)
        VALUES (rec.player_id, rec.planet_id, p_date, rec.stat_key, rec.stat_value)
        ON CONFLICT (player_id, date, stat_key)
        DO UPDATE SET stat_value = EXCLUDED.stat_value;
    END LOOP;

    -- Reset daily counters in player_stats (keep cumulative stats)
    FOR rec IN SELECT player_id, stat_key FROM player_stats
               WHERE stat_key LIKE 'daily_%'
    LOOP
        UPDATE player_stats
        SET stat_value = 0, updated_at = NOW()
        WHERE player_id = rec.player_id AND stat_key = rec.stat_key;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
