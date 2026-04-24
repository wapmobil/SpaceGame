package api

import "spacegame/internal/game"

// PlayerRequest is the request body for player registration.
type PlayerRequest struct {
	Name string `json:"name"`
}

// PlayerResponse is the response body for player registration.
type PlayerResponse struct {
	ID        string `json:"id"`
	AuthToken string `json:"auth_token"`
	Name      string `json:"name"`
}

// PlanetResponse represents a planet in API responses.
type PlanetResponse struct {
	ID        string                 `json:"id"`
	PlayerID  string                 `json:"player_id"`
	Name      string                 `json:"name"`
	Level     int                    `json:"level"`
	Resources map[string]interface{} `json:"resources"`
}

// CreatePlanetRequest is the request body for creating a planet.
type CreatePlanetRequest struct {
	Name string `json:"name"`
}

// StartResearchRequest is the request body for starting research.
type StartResearchRequest struct {
	TechID string `json:"tech_id"`
}

// BuildShipRequest is the request body for building a ship.
type BuildShipRequest struct {
	ShipType string `json:"ship_type"`
}

// BuildBuildingRequest is the request body for building a structure.
type BuildBuildingRequest struct {
	Type string `json:"type"`
}

// FleetResponse represents fleet state in API responses.
type FleetResponse struct {
	Ships            map[string]interface{} `json:"ships"`
	TotalShips       int                    `json:"total_ships"`
	TotalSlots       int                    `json:"total_slots"`
	MaxSlots         int                    `json:"max_slots"`
	TotalCargo       float64                `json:"total_cargo"`
	TotalEnergy      float64                `json:"total_energy"`
	TotalDamage      float64                `json:"total_damage"`
	TotalHP          float64                `json:"total_hp"`
	ShipyardLevel    int                    `json:"shipyard_level"`
	ShipyardQueueLen int                    `json:"shipyard_queue_len"`
	ShipyardProgress float64                `json:"shipyard_progress"`
}

// ShipTypeResponse represents a ship type in API responses.
type ShipTypeResponse struct {
	TypeID       string  `json:"type_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Slots        int     `json:"slots"`
	Cargo        float64 `json:"cargo"`
	Energy       float64 `json:"energy"`
	HP           float64 `json:"hp"`
	Armor        float64 `json:"armor"`
	WeaponMinDmg float64 `json:"weapon_min_damage"`
	WeaponMaxDmg float64 `json:"weapon_max_damage"`
	Cost         Cost    `json:"cost"`
	BuildTime    float64 `json:"build_time"`
	MinShipyard  int     `json:"min_shipyard_level"`
	CanBuild     bool    `json:"can_build"`
}

// Cost represents ship build costs.
type Cost struct {
	Food       float64 `json:"food"`
	Composite  float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents   float64 `json:"reagents"`
	Money      float64 `json:"money"`
}

// StartExpeditionRequest is the request body for starting an expedition.
type StartExpeditionRequest struct {
	ExpeditionType string  `json:"expedition_type"`
	Target         string  `json:"target"`
	Duration       float64 `json:"duration"`
	ShipTypes      []string `json:"ship_types"` // ship types to include in expedition
	ShipCounts     []int    `json:"ship_counts"` // counts for each ship type
}

// ExpeditionActionRequest is the request body for expedition actions.
type ExpeditionActionRequest struct {
	Action string `json:"action"` // "loot", "attack", "wait", "leave"
}

// ExpeditionResponse represents an expedition in API responses.
type ExpeditionResponse struct {
	ID             string                   `json:"id"`
	PlanetID       string                   `json:"planet_id"`
	Target         string                   `json:"target"`
	Progress       float64                  `json:"progress"`
	Status         string                   `json:"status"`
	ExpeditionType string                   `json:"expedition_type"`
	Duration       float64                  `json:"duration"`
	ElapsedTime    float64                  `json:"elapsed_time"`
	FleetShips     map[string]interface{}   `json:"fleet_ships"`
	FleetTotal     int                      `json:"fleet_total"`
	FleetCargo     float64                  `json:"fleet_cargo"`
	FleetEnergy    float64                  `json:"fleet_energy"`
	FleetDamage    float64                  `json:"fleet_damage"`
	DiscoveredNPC  *NPCPlanetResponse       `json:"discovered_npc,omitempty"`
	Actions        []ExpeditionActionResp   `json:"actions"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
}

// NPCPlanetResponse represents a discovered NPC planet.
type NPCPlanetResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Resources       map[string]float64     `json:"resources"`
	TotalResources  float64                `json:"total_resources"`
	HasCombat       bool                   `json:"has_combat"`
	FleetStrength   float64                `json:"fleet_strength"`
	EnemyFleet      map[string]interface{} `json:"enemy_fleet,omitempty"`
}

// ExpeditionActionResp represents an available expedition action.
type ExpeditionActionResp struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Label    string `json:"label"`
	Required string `json:"required,omitempty"`
}

// ExpeditionsListResponse is the response for listing expeditions.
type ExpeditionsListResponse struct {
	Expeditions      []ExpeditionResponse `json:"expeditions"`
	ActiveCount      int                  `json:"active_count"`
	MaxExpeditions   int                  `json:"max_expeditions"`
	CanStartNew      bool                 `json:"can_start_new"`
	ExpeditionsUnlocked bool              `json:"expeditions_unlocked"`
}

// CreateMarketOrderRequest is the request body for creating a market order.
type CreateMarketOrderRequest struct {
	Resource  string  `json:"resource"`
	OrderType string  `json:"order_type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	IsPrivate bool    `json:"is_private"`
}

// MarketOrderResponse represents a market order in API responses.
type MarketOrderResponse struct {
	ID               string                 `json:"id"`
	PlanetID         string                 `json:"planet_id"`
	PlayerID         string                 `json:"player_id"`
	Resource         string                 `json:"resource"`
	OrderType        string                 `json:"order_type"`
	Amount           float64                `json:"amount"`
	Price            float64                `json:"price"`
	IsPrivate        bool                   `json:"is_private"`
	Link             string                 `json:"link,omitempty"`
	Status           string                 `json:"status"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	ReservedResources map[string]float64    `json:"reserved_resources,omitempty"`
}

// MiningMoveRequest is the request body for moving in a mining session.
type MiningMoveRequest struct {
	Direction string `json:"direction"` // "up", "down", "left", "right"
	Slide     bool   `json:"slide"`     // if true, slide to wall
}

// MiningStateResponse represents the current state of a mining session.
type MiningStateResponse struct {
	SessionID      string     `json:"session_id"`
	PlanetID       string     `json:"planet_id"`
	PlayerID       string     `json:"player_id"`
	Maze           [][]string `json:"maze"`
	PlayerX        int        `json:"player_x"`
	PlayerY        int        `json:"player_y"`
	PlayerHP       int        `json:"player_hp"`
	PlayerMaxHP    int        `json:"player_max_hp"`
	PlayerBombs    int        `json:"player_bombs"`
	MoneyCollected float64    `json:"money_collected"`
	Status         string     `json:"status"`
	ExitX          int        `json:"exit_x"`
	ExitY          int        `json:"exit_y"`
	BaseLevel      int        `json:"base_level"`
	Monsters       []MonsterResponse `json:"monsters"`
	AvailableMoves []string    `json:"available_moves"`
	StartTime      string    `json:"start_time"`
	CompletedAt    string    `json:"completed_at,omitempty"`
}

// MonsterResponse represents a monster in API responses.
type MonsterResponse struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Icon      string  `json:"icon"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	HP        int     `json:"hp"`
	MaxHP     int     `json:"max_hp"`
	Damage    int     `json:"damage"`
	Reward    float64 `json:"reward"`
	Alive     bool    `json:"alive"`
}

// MiningMoveResponse represents the result of a move action.
type MiningMoveResponse struct {
	Success        bool            `json:"success"`
	Message        string          `json:"message,omitempty"`
	Maze           [][]string      `json:"maze"`
	PlayerX        int             `json:"player_x"`
	PlayerY        int             `json:"player_y"`
	PlayerHP       int             `json:"player_hp"`
	PlayerBombs    int             `json:"player_bombs"`
	MoneyCollected float64         `json:"money_collected"`
	Encounter      *EncounterResponse `json:"encounter,omitempty"`
	GameEnded      bool            `json:"game_ended"`
	EndReason      string          `json:"end_reason,omitempty"`
}

// EncounterResponse represents a monster encounter.
type EncounterResponse struct {
	MonsterID   string  `json:"monster_id"`
	MonsterName string  `json:"monster_name"`
	MonsterIcon string  `json:"monster_icon"`
	Damage      int     `json:"damage"`
	Reward      float64 `json:"reward"`
	Killed      bool    `json:"killed"`
}

// MiningStartResponse is the response for starting a mining session.
type MiningStartResponse struct {
	PlanetID       string     `json:"planet_id"`
	Status         string     `json:"status"`
	SessionID      string     `json:"session_id"`
	Maze           [][]string `json:"maze"`
	PlayerX        int        `json:"player_x"`
	PlayerY        int        `json:"player_y"`
	PlayerHP       int        `json:"player_hp"`
	PlayerMaxHP    int        `json:"player_max_hp"`
	PlayerBombs    int        `json:"player_bombs"`
	MoneyCollected float64    `json:"money_collected"`
	ExitX          int        `json:"exit_x"`
	ExitY          int        `json:"exit_y"`
	BaseLevel      int        `json:"base_level"`
	Monsters       []MonsterResponse `json:"monsters"`
	AvailableMoves []string   `json:"available_moves"`
}



// BuildingDetail represents a building with all computed data for the frontend.
type BuildingDetail struct {
	Type           string      `json:"type"`
	Level          int         `json:"level"`
	BuildProgress  float64     `json:"build_progress"`
	Enabled        bool        `json:"enabled"`
	BuildTime      float64     `json:"build_time"`
	Cost           CostDetail  `json:"cost"`
	NextCost       CostDetail  `json:"next_cost"`
	Production     ProdDetail  `json:"production"`
	NextProduction ProdDetail  `json:"next_production"`
	Deltas         ProdDetail  `json:"deltas"`
}

// BuildingCostDetail represents cost + production info for unbuilt buildings.
type BuildingCostDetail struct {
	Cost           CostDetail `json:"cost"`
	Production     ProdDetail `json:"production"`
	NextProduction ProdDetail `json:"next_production"`
	Deltas         ProdDetail `json:"deltas"`
}

// CostDetail represents build costs for the API.
type CostDetail struct {
	Food  float64 `json:"food"`
	Iron  float64 `json:"iron"`
	Money float64 `json:"money"`
}

// ProdDetail represents per-tick resource production for the API.
type ProdDetail struct {
	Food       float64 `json:"food"`
	Iron       float64 `json:"iron"`
	Composite  float64 `json:"composite"`
	Mechanisms float64 `json:"mechanisms"`
	Reagents   float64 `json:"reagents"`
	Energy     float64 `json:"energy"`
	Money      float64 `json:"money"`
	AlienTech  float64 `json:"alien_tech"`
	EnergyNet  float64 `json:"energy_net"`
}

// EnergyBufferDetail represents the energy buffer state.
type EnergyBufferDetail struct {
	Value   float64 `json:"value"`
	Max     float64 `json:"max"`
	Deficit bool    `json:"deficit"`
}

// BuildDetailsResponse is the response for GET /api/planets/{id}/build-details.
type BuildDetailsResponse struct {
	Resources          PlanetResources            `json:"resources"`
	EnergyBuffer       EnergyBufferDetail         `json:"energy_buffer"`
	Buildings          []BuildingDetail           `json:"buildings"`
	EnergyBalance      EnergyBalanceDetail        `json:"energy_balance"`
	ResourceProduction ProdDetail                 `json:"production"`
	ActiveConstruction int                        `json:"active_constructions"`
	MaxConstruction    int                        `json:"max_constructions"`
	BaseOperational    bool                       `json:"base_operational"`
	CanResearch        bool                       `json:"can_research"`
	CanExpedition      bool                       `json:"can_expedition"`
	CanMining          bool                       `json:"can_mining"`
	BuildingCosts      map[string]BuildingCostDetail `json:"building_costs"`
	ResearchUnlocks    string                     `json:"research_unlocks"`
}

// EnergyBalanceDetail represents energy production and consumption.
type EnergyBalanceDetail struct {
	Production  float64 `json:"production"`
	Consumption float64 `json:"consumption"`
	Net         float64 `json:"net"`
}

// PlanetResources is an alias for the game package's PlanetResources.
type PlanetResources = game.PlanetResources
