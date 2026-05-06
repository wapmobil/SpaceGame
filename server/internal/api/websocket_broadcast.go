package api

import (
	"encoding/json"
	"fmt"
)

// WSBroadcastService provides broadcast functionality for game events.
type WSBroadcastService struct {
	cm *WSConnectionManager
}

// NewWSBroadcastService creates a new broadcast service.
func NewWSBroadcastService(cm *WSConnectionManager) *WSBroadcastService {
	return &WSBroadcastService{cm: cm}
}

// BroadcastStateUpdate sends a planet state update to the owning player.
func (bs *WSBroadcastService) BroadcastStateUpdate(playerID, planetID string, state map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "state_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","state":%s}`, planetID, toJSON(state))),
	})
}

// BroadcastNotification sends a notification to a player.
func (bs *WSBroadcastService) BroadcastNotification(playerID, message, notifType string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"%s","type":"%s"}`, escapeJSON(message), escapeJSON(notifType))),
	})
}

// BroadcastExpeditionUpdate sends an expedition update to the owning player.
func (bs *WSBroadcastService) BroadcastExpeditionUpdate(playerID, expeditionID string, progress float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "space_expedition_update",
		Data: json.RawMessage(fmt.Sprintf(`{"expedition_id":"%s","progress":%.4f}`, expeditionID, progress)),
	})
}

// BroadcastMarketUpdate sends a market update to the owning player.
func (bs *WSBroadcastService) BroadcastMarketUpdate(playerID, orderID, status string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "market_update",
		Data: json.RawMessage(fmt.Sprintf(`{"order_id":"%s","status":"%s"}`, orderID, status)),
	})
}

// BroadcastBuildShip sends a ship build notification.
func (bs *WSBroadcastService) BroadcastBuildShip(playerID, planetID, shipType string, queued bool) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "build",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","ship_type":"%s","queued":%t}`, planetID, shipType, queued)),
	})
}

// BroadcastResearchUpdate sends a research update.
func (bs *WSBroadcastService) BroadcastResearchUpdate(playerID, planetID, techID string, progress float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "research_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","tech_id":"%s","progress":%.4f}`, planetID, techID, progress)),
	})
}

// BroadcastBuildingUpdate sends a building update.
func (bs *WSBroadcastService) BroadcastBuildingUpdate(playerID, planetID, building string, level int) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "building_update",
		Data: json.RawMessage(fmt.Sprintf(`{"planet_id":"%s","building":"%s","level":%d}`, planetID, building, level)),
	})
}

// BroadcastDrillUpdate sends a drill mini-game update to the owning player.
func (bs *WSBroadcastService) BroadcastDrillUpdate(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "drill_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastGardenBedUpdate sends a garden bed update to the owning player.
func (bs *WSBroadcastService) BroadcastGardenBedUpdate(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "garden_bed_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastRandomEvent sends a random event notification.
func (bs *WSBroadcastService) BroadcastRandomEvent(playerID string, event map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "random_event",
		Data: json.RawMessage(toJSON(event)),
	})
}

// BroadcastPlayerKilled sends a player death notification.
func (bs *WSBroadcastService) BroadcastPlayerKilled(playerID string, opponent string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"Your fleet was defeated by %s","type":"combat"}`, escapeJSON(opponent))),
	})
}

// BroadcastExpeditionDiscovered sends an expedition discovery notification.
func (bs *WSBroadcastService) BroadcastExpeditionDiscovered(playerID, expeditionID, npcName, npcType string) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "notification",
		Data: json.RawMessage(fmt.Sprintf(`{"message":"Expedition discovered %s (%s)","type":"space_discovery","expedition_id":"%s"}`, escapeJSON(npcName), escapeJSON(npcType), expeditionID)),
	})
}

// BroadcastPlanetSurveyUpdate sends a planet survey update to the owning player.
func (bs *WSBroadcastService) BroadcastPlanetSurveyUpdate(playerID, planetID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "planet_survey_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastLocationUpdate sends a location update to the owning player.
func (bs *WSBroadcastService) BroadcastLocationUpdate(playerID, planetID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "location_update",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastExpeditionEvent sends an expedition event update to the owning player.
func (bs *WSBroadcastService) BroadcastExpeditionEvent(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "expedition_event",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastExpeditionComplete sends an expedition completion notification.
func (bs *WSBroadcastService) BroadcastExpeditionComplete(playerID string, data map[string]interface{}) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "expedition_complete",
		Data: json.RawMessage(toJSON(data)),
	})
}

// BroadcastMarketOrderFilled sends a market order filled notification.
func (bs *WSBroadcastService) BroadcastMarketOrderFilled(playerID, orderID, resource string, amount, price float64) {
	bs.cm.SendToPlayer(playerID, WSMessage{
		Type: "market_update",
		Data: json.RawMessage(fmt.Sprintf(`{"order_id":"%s","status":"filled","resource":"%s","amount":%.0f,"price":%.0f}`, orderID, resource, amount, price)),
	})
}

// broadcastToPlayer sends a message to a player, queuing if disconnected.
func (bs *WSBroadcastService) broadcastToPlayer(playerID string, msg WSMessage) {
	bs.cm.SendToPlayer(playerID, msg)
}

// broadcastToAll sends a message to all connected players.
func (bs *WSBroadcastService) broadcastToAll(msg WSMessage) {
	bs.cm.BroadcastToAll(msg)
}

// broadcastToPlanet sends a message to all players with a specific planet.
func (bs *WSBroadcastService) broadcastToPlanet(planetID string, msg WSMessage) {
	bs.cm.SendToPlanet(planetID, msg)
}

// Helper functions

func escapeJSON(s string) string {
	s = replaceAll(s, `\`, `\\`)
	s = replaceAll(s, `"`, `\"`)
	s = replaceAll(s, "\n", `\n`)
	s = replaceAll(s, "\r", `\r`)
	s = replaceAll(s, "\t", `\t`)
	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
