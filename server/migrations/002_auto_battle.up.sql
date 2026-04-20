-- Drop old battles table and recreate with simplified schema
DROP TABLE IF EXISTS battles;

CREATE TABLE IF NOT EXISTS battles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attacker_planet_id UUID NOT NULL REFERENCES planets(id),
    defender_planet_id UUID NOT NULL REFERENCES planets(id),
    defender_is_npc BOOLEAN NOT NULL DEFAULT FALSE,
    result TEXT NOT NULL DEFAULT 'pending' CHECK (result IN ('pending','attacker_win','defender_win','draw')),
    attacker_loss JSONB NOT NULL DEFAULT '{}',
    defender_loss JSONB NOT NULL DEFAULT '{}',
    loot JSONB NOT NULL DEFAULT '{}',
    rounds INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_battles_updated_at BEFORE UPDATE ON battles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
