package api

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
