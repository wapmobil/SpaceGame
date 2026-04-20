-- Revert battles table to original schema
DROP TABLE IF EXISTS battles;

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

CREATE TRIGGER update_battles_updated_at BEFORE UPDATE ON battles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
