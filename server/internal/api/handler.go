package api

// handler.go — all handlers have been split into domain-specific files:
//   auth_handlers.go    — health, register, login, list_planets, create_planet, get_planet
//   building_handlers.go — get_buildings, build_building, confirm_building, toggle_building, get_build_details
//   research_handlers.go — get_research, start_research
//   fleet_handlers.go    — get_fleet, build_ship, get_available_ships
//   expedition_handlers.go — create_expedition, get_expeditions, expedition_action
//   market_handlers.go   — create_order, get_my_orders, get_global_market, delete_order, sell_food
//   drill_handlers.go    — start_drill, drill_command, drill_chunk, complete_drill, destroy_drill, cleanup_drill
//   farm_handlers.go     — get_farm, farm_action
//   other_handlers.go    — get_ratings, get_stats, resolve_event
