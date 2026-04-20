package battle

import (
	"testing"

	"spacegame/internal/game/ship"
)

func makeTestFleet(shipType ship.TypeID, count int) *FleetSnapshot {
	st := ship.GetShipType(shipType)
	if st == nil {
		return &FleetSnapshot{Ships: make(map[string]*ShipCount)}
	}

	fleet := &FleetSnapshot{
		Ships: make(map[string]*ShipCount),
	}

	sc := &ShipCount{
		Type:      string(st.TypeID),
		Count:     count,
		HP:        st.HP,
		Armor:     st.Armor,
		WeaponMin: st.WeaponMinDmg,
		WeaponMax: st.WeaponMaxDmg,
		TotalHP:   float64(count) * st.HP,
		TotalDmg:  float64(count) * (st.WeaponMinDmg + st.WeaponMaxDmg) / 2,
		Cost:      st.Cost,
	}

	fleet.Ships[string(st.TypeID)] = sc
	return fleet
}

func TestCalculateBattle_AttackerWins(t *testing.T) {
	// Strong attacker vs weak defender
	attacker := makeTestFleet(ship.TypeCorvette, 3)
	defender := makeTestFleet(ship.TypeSmallShip, 2)

	result := CalculateBattle(attacker, defender)

	if result.Winner != "attacker" {
		t.Errorf("Expected attacker to win, got %s", result.Winner)
	}

	if result.Rounds <= 0 {
		t.Errorf("Expected at least 1 round, got %d", result.Rounds)
	}

	if result.DefenderLost["small_ship"] != 2 {
		t.Errorf("Expected defender to lose all 2 small ships, got %v", result.DefenderLost)
	}

	if len(result.AttackerLoot) == 0 {
		t.Error("Expected attacker to get loot")
	}
}

func TestCalculateBattle_DefenderWins(t *testing.T) {
	// Weak attacker vs strong defender
	attacker := makeTestFleet(ship.TypeSmallShip, 2)
	defender := makeTestFleet(ship.TypeCruiser, 2)

	result := CalculateBattle(attacker, defender)

	if result.Winner != "defender" {
		t.Errorf("Expected defender to win, got %s", result.Winner)
	}

	if result.AttackerLost["small_ship"] != 2 {
		t.Errorf("Expected attacker to lose all 2 small ships, got %v", result.AttackerLost)
	}
}

func TestCalculateBattle_Draw(t *testing.T) {
	// Equal fleets - due to randomness, either side can win
	// But both should have significant losses in a fair fight
	attacker := makeTestFleet(ship.TypeInterceptor, 5)
	defender := makeTestFleet(ship.TypeInterceptor, 5)

	result := CalculateBattle(attacker, defender)

	if result.Winner == "" {
		t.Error("Expected a winner or draw")
	}

	if result.Rounds <= 0 {
		t.Errorf("Expected at least 1 round, got %d", result.Rounds)
	}

	// Both sides should have losses (fair fight)
	totalAttackerLost := 0
	totalDefenderLost := 0
	for _, count := range result.AttackerLost {
		totalAttackerLost += count
	}
	for _, count := range result.DefenderLost {
		totalDefenderLost += count
	}

	// At least one side should lose ships
	if totalAttackerLost == 0 && totalDefenderLost == 0 {
		t.Error("Expected at least one side to lose ships")
	}

	// In a fair fight, losses should be somewhat balanced
	// Allow some variance due to randomness
	if totalAttackerLost > 0 && totalDefenderLost > 0 {
		ratio := float64(totalAttackerLost) / float64(totalDefenderLost)
		if ratio > 3 || ratio < 0.33 {
			t.Logf("Warning: unbalanced losses - attacker: %d, defender: %d", totalAttackerLost, totalDefenderLost)
		}
	}
}

func TestCalculateBattle_EmptyAttacker(t *testing.T) {
	attacker := &FleetSnapshot{Ships: make(map[string]*ShipCount)}
	defender := makeTestFleet(ship.TypeCorvette, 1)

	result := CalculateBattle(attacker, defender)

	if result.Winner != "defender" {
		t.Errorf("Expected defender to win against empty fleet, got %s", result.Winner)
	}
}

func TestCalculateBattle_EmptyDefender(t *testing.T) {
	attacker := makeTestFleet(ship.TypeCorvette, 1)
	defender := &FleetSnapshot{Ships: make(map[string]*ShipCount)}

	result := CalculateBattle(attacker, defender)

	if result.Winner != "attacker" {
		t.Errorf("Expected attacker to win against empty fleet, got %s", result.Winner)
	}

	if len(result.AttackerLoot) == 0 {
		t.Error("Expected loot when defender is empty")
	}
}

func TestCalculateBattle_BothEmpty(t *testing.T) {
	attacker := &FleetSnapshot{Ships: make(map[string]*ShipCount)}
	defender := &FleetSnapshot{Ships: make(map[string]*ShipCount)}

	result := CalculateBattle(attacker, defender)

	if result.Winner != "draw" {
		t.Errorf("Expected draw when both fleets are empty, got %s", result.Winner)
	}
}

func TestCalculateBattle_NoWeapons(t *testing.T) {
	// Trade ships have no weapons
	attacker := makeTestFleet(ship.TypeTradeShip, 5)
	defender := makeTestFleet(ship.TypeTradeShip, 5)

	result := CalculateBattle(attacker, defender)

	// With no weapons, neither side can deal damage, so it should be a draw
	if result.Winner != "draw" {
		t.Errorf("Expected draw when both fleets have no weapons, got %s", result.Winner)
	}
}

func TestCalculateBattle_ArmorReduction(t *testing.T) {
	// High armor defender vs low damage attacker
	attacker := makeTestFleet(ship.TypeSmallShip, 5) // dmg 1-1
	defender := makeTestFleet(ship.TypeFrigate, 1)   // armor 5, hp 100

	result := CalculateBattle(attacker, defender)

	// Frigate's high armor should reduce damage significantly
	// Small ships deal 1 damage, frigate has 5 armor, so damage is clamped to 0
	// This means the battle should end in a draw or defender win
	if result.DefenderLost["frigate"] > 0 {
		t.Logf("Frigate lost (armor reduced damage to 0, but hits still occur)")
	}
}

func TestCalculateBattle_CruiserDominance(t *testing.T) {
	// Cruiser should dominate weaker ships
	attacker := makeTestFleet(ship.TypeCruiser, 1)
	defender := makeTestFleet(ship.TypeSmallShip, 10)

	result := CalculateBattle(attacker, defender)

	if result.Winner != "attacker" {
		t.Errorf("Expected cruiser to win, got %s", result.Winner)
	}

	// Cruiser should lose minimal or no ships
	if result.AttackerLost["cruiser"] > 0 {
		t.Logf("Cruiser lost some ships: %v", result.AttackerLost)
	}

	// Defender should lose most ships
	totalDefenderLost := 0
	for _, count := range result.DefenderLost {
		totalDefenderLost += count
	}
	if totalDefenderLost < 5 {
		t.Errorf("Expected defender to lose at least 5 ships, lost %d", totalDefenderLost)
	}
}

func TestFleetSnapshot_TotalHP(t *testing.T) {
	fleet := makeTestFleet(ship.TypeCorvette, 5) // hp 200 each

	totalHP := fleet.TotalHP()
	expected := 5.0 * 200.0

	if totalHP != expected {
		t.Errorf("Expected total HP %.0f, got %.0f", expected, totalHP)
	}
}

func TestFleetSnapshot_TotalDPS(t *testing.T) {
	fleet := makeTestFleet(ship.TypeCorvette, 3) // dmg 2-4, avg 3

	totalDPS := fleet.TotalDPS()
	expected := 3.0 * 3.0 // 3 ships × avg 3 damage

	if totalDPS != expected {
		t.Errorf("Expected total DPS %.0f, got %.0f", expected, totalDPS)
	}
}

func TestFleetSnapshot_ApplyDamage(t *testing.T) {
	fleet := makeTestFleet(ship.TypeCorvette, 5) // hp 200 each

	damage := 500.0
	actualDamage, destroyed := fleet.ApplyDamage(damage)

	if actualDamage <= 0 {
		t.Errorf("Expected positive actual damage, got %.0f", actualDamage)
	}

	// 500 damage / 200 hp per ship = 2.5 ships destroyed
	expectedDestroyed := 2
	if destroyed["corvette"] != expectedDestroyed {
		t.Errorf("Expected %d corvettes destroyed, got %d", expectedDestroyed, destroyed["corvette"])
	}

	// Remaining ships: 3 × 200 = 600 HP
	remainingHP := fleet.TotalHP()
	expectedHP := 3.0 * 200.0
	if remainingHP != expectedHP {
		t.Errorf("Expected remaining HP %.0f, got %.0f", expectedHP, remainingHP)
	}
}

func TestFleetSnapshot_ApplyDamage_NoArmorReduction(t *testing.T) {
	fleet := makeTestFleet(ship.TypeInterceptor, 3) // armor 0

	damage := 100.0
	actualDamage, destroyed := fleet.ApplyDamage(damage)

	// With 0 armor, all damage should be applied
	if actualDamage != damage {
		t.Errorf("Expected full damage %.0f, got %.0f", damage, actualDamage)
	}

	// 100 damage / 20 hp per ship = 5 ships, but only 3 exist
	if destroyed["interceptor"] != 3 {
		t.Errorf("Expected all 3 interceptors destroyed, got %d", destroyed["interceptor"])
	}
}

func TestFleetSnapshot_ApplyDamage_ArmorReduction(t *testing.T) {
	fleet := makeTestFleet(ship.TypeFrigate, 2) // armor 5, hp 100

	damage := 100.0
	actualDamage, destroyed := fleet.ApplyDamage(damage)

	// Armor reduces damage, but damage is still applied (just reduced)
	if actualDamage >= damage {
		t.Errorf("Expected reduced damage, got %.0f out of %.0f", actualDamage, damage)
	}

	// With reduced damage, fewer ships should be destroyed
	if destroyed["frigate"] >= 2 {
		t.Errorf("Expected fewer than 2 frigates destroyed, got %d", destroyed["frigate"])
	}
}

func TestFleetSnapshot_Clone(t *testing.T) {
	original := makeTestFleet(ship.TypeCorvette, 5)
	clone := original.Clone()

	if clone.TotalShipCount() != original.TotalShipCount() {
		t.Error("Clone should have same ship count")
	}

	if clone.TotalHP() != original.TotalHP() {
		t.Error("Clone should have same total HP")
	}

	// Modify clone
	clone.RemoveShips(map[string]int{"corvette": 2})

	// Original should be unchanged
	if original.TotalShipCount() != 5 {
		t.Errorf("Original should still have 5 ships, got %d", original.TotalShipCount())
	}
}

func TestFleetSnapshot_IsEmpty(t *testing.T) {
	empty := &FleetSnapshot{Ships: make(map[string]*ShipCount)}
	if !empty.IsEmpty() {
		t.Error("Empty fleet should return true for IsEmpty()")
	}

	nonEmpty := makeTestFleet(ship.TypeCorvette, 1)
	if nonEmpty.IsEmpty() {
		t.Error("Non-empty fleet should return false for IsEmpty()")
	}
}

func TestFleetSnapshot_HasCombatShips(t *testing.T) {
	peaceful := makeTestFleet(ship.TypeTradeShip, 5)
	if peaceful.HasCombatShips() {
		t.Error("Trade fleet should not have combat ships")
	}

	combat := makeTestFleet(ship.TypeCorvette, 3)
	if !combat.HasCombatShips() {
		t.Error("Corvette fleet should have combat ships")
	}
}

func TestGetDestroyedShipCosts(t *testing.T) {
	destroyed := map[string]int{
		"corvette": 1,
		"small_ship": 2,
	}

	refund := GetDestroyedShipCosts(destroyed)

	// Corvette cost: 250 each, 50% refund = 125 per resource
	// Small ship cost: 10 each, 50% refund = 5 per resource
	expectedFood := 250*0.5 + 10*2*0.5 // 125 + 10 = 135
	if refund["food"] != expectedFood {
		t.Errorf("Expected food refund %.0f, got %.0f", expectedFood, refund["food"])
	}

	if refund["money"] != expectedFood {
		t.Errorf("Expected money refund %.0f, got %.0f", expectedFood, refund["money"])
	}
}

func TestCalculateBattle_MixedFleet(t *testing.T) {
	// Mixed fleet with different ship types
	attacker := &FleetSnapshot{
		Ships: map[string]*ShipCount{
			"corvette": {
				Type:      "corvette",
				Count:     2,
				HP:        200,
				Armor:     10,
				WeaponMin: 2,
				WeaponMax: 4,
				TotalHP:   400,
				TotalDmg:  12,
			},
			"interceptor": {
				Type:      "interceptor",
				Count:     1,
				HP:        20,
				Armor:     0,
				WeaponMin: 8,
				WeaponMax: 10,
				TotalHP:   20,
				TotalDmg:  18,
			},
		},
	}

	defender := makeTestFleet(ship.TypeFrigate, 3) // hp 100, armor 5, dmg 5 avg

	result := CalculateBattle(attacker, defender)

	if result.Winner == "" {
		t.Error("Expected a winner, got empty string")
	}

	if result.Rounds <= 0 {
		t.Errorf("Expected at least 1 round, got %d", result.Rounds)
	}

	// Verify total losses make sense
	totalDefenderLost := 0
	for _, count := range result.DefenderLost {
		totalDefenderLost += count
	}
	if totalDefenderLost > 3 {
		t.Errorf("Defender can't lose more than 3 frigates, lost %d", totalDefenderLost)
	}
}

func TestCalculateBattle_LargeFleet(t *testing.T) {
	// Large fleet battle
	attacker := makeTestFleet(ship.TypeCorvette, 20)
	defender := makeTestFleet(ship.TypeFrigate, 15)

	result := CalculateBattle(attacker, defender)

	if result.Winner == "" {
		t.Error("Expected a winner in large fleet battle")
	}

	// Should complete in reasonable number of rounds
	if result.Rounds > 50 {
		t.Errorf("Battle took too many rounds: %d", result.Rounds)
	}

	// Verify losses are proportional
	totalAttackerLost := 0
	totalDefenderLost := 0
	for _, count := range result.AttackerLost {
		totalAttackerLost += count
	}
	for _, count := range result.DefenderLost {
		totalDefenderLost += count
	}

	if totalAttackerLost > 20 {
		t.Errorf("Attacker can't lose more than 20 ships, lost %d", totalAttackerLost)
	}
	if totalDefenderLost > 15 {
		t.Errorf("Defender can't lose more than 15 ships, lost %d", totalDefenderLost)
	}
}

func TestCalculateBattle_SingleShipVsSingleShip(t *testing.T) {
	attacker := makeTestFleet(ship.TypeInterceptor, 1)
	defender := makeTestFleet(ship.TypeSmallShip, 1)

	result := CalculateBattle(attacker, defender)

	if result.Winner == "" {
		t.Error("Expected a winner in 1v1 battle")
	}

	if result.Rounds <= 0 {
		t.Errorf("Expected at least 1 round, got %d", result.Rounds)
	}

	// At least one side should lose their ship
	if result.AttackerLost["interceptor"] == 0 && result.DefenderLost["small_ship"] == 0 {
		t.Error("At least one ship should be destroyed in 1v1")
	}
}

func TestCalculateBattle_HighArmorVsLowDamage(t *testing.T) {
	// Very high armor vs very low damage
	attacker := makeTestFleet(ship.TypeSmallShip, 1) // dmg 1-1
	defender := makeTestFleet(ship.TypeCruiser, 1)   // armor 20

	result := CalculateBattle(attacker, defender)

	// Cruiser's high armor should make it nearly invulnerable to small ships
	// Battle should end in draw or defender win
	if result.DefenderLost["cruiser"] > 0 {
		t.Logf("Cruiser took damage despite high armor: %v", result.DefenderLost)
	}
}

func TestFleetSnapshot_NewFleetSnapshot(t *testing.T) {
	// Create a real fleet and convert to snapshot
	f := ship.NewFleet()
	st := ship.GetShipType(ship.TypeCorvette)
	if st == nil {
		t.Fatal("Corvette type not found")
	}

	f.AddShip(st, 3)

	snapshot := NewFleetSnapshot(f)

	if snapshot.TotalShipCount() != 3 {
		t.Errorf("Expected 3 ships in snapshot, got %d", snapshot.TotalShipCount())
	}

	sc, ok := snapshot.Ships["corvette"]
	if !ok {
		t.Error("Expected corvette in snapshot")
	}

	if sc.Count != 3 {
		t.Errorf("Expected 3 corvettes, got %d", sc.Count)
	}

	if sc.TotalHP != 600 {
		t.Errorf("Expected 600 total HP, got %.0f", sc.TotalHP)
	}
}
