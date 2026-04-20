-- Enhance market_orders table with additional fields for marketplace Phase 7
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS player_id UUID REFERENCES players(id);
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS is_private BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS link TEXT;
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','filled','cancelled','expired'));
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS reserved_food REAL NOT NULL DEFAULT 0;
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS reserved_composite REAL NOT NULL DEFAULT 0;
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS reserved_mechanisms REAL NOT NULL DEFAULT 0;
ALTER TABLE market_orders ADD COLUMN IF NOT EXISTS reserved_reagents REAL NOT NULL DEFAULT 0;

-- Create npc_traders table
CREATE TABLE IF NOT EXISTS npc_traders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    planet_id UUID REFERENCES planets(id) ON DELETE SET NULL,
    order_id UUID REFERENCES market_orders(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_npc_traders_updated_at BEFORE UPDATE ON npc_traders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_market_orders_player_id ON market_orders(player_id);
CREATE INDEX IF NOT EXISTS idx_market_orders_status ON market_orders(status);
CREATE INDEX IF NOT EXISTS idx_market_orders_resource ON market_orders(resource);
CREATE INDEX IF NOT EXISTS idx_market_orders_type ON market_orders(type);
