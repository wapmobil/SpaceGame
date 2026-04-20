package ship

import (
	"testing"
)

func TestShipTypesExist(t *testing.T) {
	types := AllShipTypes()
	if len(types) != 6 {
		t.Errorf("expected 6 ship types, got %d", len(types))
	}
}

func TestTradeShipCost(t *testing.T) {
	st := GetShipType(TypeTradeShip)
	if st == nil {
		t.Fatal("TradeShip not found")
	}
	if st.Cost.Food != 100 {
		t.Errorf("expected trade ship food cost 100, got %f", st.Cost.Food)
	}
	if st.Cost.Money != 100 {
		t.Errorf("expected trade ship money cost 100, got %f", st.Cost.Money)
	}
	if st.Slots != 2 {
		t.Errorf("expected trade ship slots 2, got %d", st.Slots)
	}
	if st.Cargo != 100 {
		t.Errorf("expected trade ship cargo 100, got %f", st.Cargo)
	}
	if st.WeaponMinDmg != 0 || st.WeaponMaxDmg != 0 {
		t.Errorf("expected trade ship to have no weapons")
	}
	if !st.IsPeaceful() {
		t.Error("expected trade ship to be peaceful")
	}
}

func TestSmallShipCost(t *testing.T) {
	st := GetShipType(TypeSmallShip)
	if st == nil {
		t.Fatal("SmallShip not found")
	}
	if st.Cost.Food != 10 {
		t.Errorf("expected small ship food cost 10, got %f", st.Cost.Food)
	}
	if st.WeaponMinDmg != 1 || st.WeaponMaxDmg != 1 {
		t.Errorf("expected small ship weapon 1-1, got %f-%f", st.WeaponMinDmg, st.WeaponMaxDmg)
	}
}

func TestInterceptorCost(t *testing.T) {
	st := GetShipType(TypeInterceptor)
	if st == nil {
		t.Fatal("Interceptor not found")
	}
	if st.WeaponMinDmg != 8 || st.WeaponMaxDmg != 10 {
		t.Errorf("expected interceptor weapon 8-10, got %f-%f", st.WeaponMinDmg, st.WeaponMaxDmg)
	}
	if st.Cargo != 0 {
		t.Errorf("expected interceptor cargo 0, got %f", st.Cargo)
	}
}

func TestCorvetteCost(t *testing.T) {
	st := GetShipType(TypeCorvette)
	if st == nil {
		t.Fatal("Corvette not found")
	}
	if st.Cost.Food != 250 {
		t.Errorf("expected corvette food cost 250, got %f", st.Cost.Food)
	}
	if st.Armor != 10 {
		t.Errorf("expected corvette armor 10, got %f", st.Armor)
	}
	if st.WeaponMinDmg != 2 || st.WeaponMaxDmg != 4 {
		t.Errorf("expected corvette weapon 2-4, got %f-%f", st.WeaponMinDmg, st.WeaponMaxDmg)
	}
}

func TestFrigateCost(t *testing.T) {
	st := GetShipType(TypeFrigate)
	if st == nil {
		t.Fatal("Frigate not found")
	}
	if st.Cost.Food != 500 {
		t.Errorf("expected frigate food cost 500, got %f", st.Cost.Food)
	}
	if st.Armor != 5 {
		t.Errorf("expected frigate armor 5, got %f", st.Armor)
	}
}

func TestCruiserCost(t *testing.T) {
	st := GetShipType(TypeCruiser)
	if st == nil {
		t.Fatal("Cruiser not found")
	}
	if st.Cost.Food != 1000 {
		t.Errorf("expected cruiser food cost 1000, got %f", st.Cost.Food)
	}
	if st.Armor != 20 {
		t.Errorf("expected cruiser armor 20, got %f", st.Armor)
	}
	if st.WeaponMinDmg != 2 || st.WeaponMaxDmg != 16 {
		t.Errorf("expected cruiser weapon 2-16, got %f-%f", st.WeaponMinDmg, st.WeaponMaxDmg)
	}
}

func TestCostCanAfford(t *testing.T) {
	st := GetShipType(TypeSmallShip)
	if !st.Cost.CanAfford(100, 100, 100, 100, 100) {
		t.Error("expected small ship to be affordable")
	}

	if st.Cost.CanAfford(5, 5, 5, 5, 5) {
		t.Error("expected small ship to not be affordable with 5 resources")
	}
}

func TestCostDeduct(t *testing.T) {
	st := GetShipType(TypeSmallShip)
	deltas := st.Cost.Deduct()
	if deltas["food"] != -10 {
		t.Errorf("expected food delta -10, got %f", deltas["food"])
	}
	if deltas["money"] != -10 {
		t.Errorf("expected money delta -10, got %f", deltas["money"])
	}
}

func TestFleetEnergy(t *testing.T) {
	fleet := NewFleet()
	st := GetShipType(TypeSmallShip)
	fleet.AddShip(st, 3)

	energy := fleet.TotalEnergyConsumption()
	expected := 10.0 * 3
	if energy != expected {
		t.Errorf("expected 30 energy for 3 small ships, got %f", energy)
	}
}

func TestFleetSlots(t *testing.T) {
	fleet := NewFleet()
	st := GetShipType(TypeCorvette)
	fleet.AddShip(st, 2)

	slots := fleet.TotalSlots()
	if slots != 8 {
		t.Errorf("expected 8 slots for 2 corvettes, got %d", slots)
	}
}

func TestFleetCargoCapacity(t *testing.T) {
	fleet := NewFleet()
	tradeShip := GetShipType(TypeTradeShip)
	fleet.AddShip(tradeShip, 2)

	cargo := fleet.TotalCargoCapacity()
	if cargo != 200 {
		t.Errorf("expected 200 cargo for 2 trade ships, got %f", cargo)
	}
}

func TestFleetMaxSlotsLimit(t *testing.T) {
	fleet := NewFleet()
	st := GetShipType(TypeCruiser)

	if !fleet.CanAddShip(st, 1, 10) {
		t.Error("expected to fit 1 cruiser in 10 slots")
	}

	fleet.AddShip(st, 2)
	if fleet.CanAddShip(st, 1, 10) {
		t.Error("expected to not fit another cruiser with 12 slots used")
	}
}

func TestFleetShipCount(t *testing.T) {
	fleet := NewFleet()
	small := GetShipType(TypeSmallShip)
	fleet.AddShip(small, 5)
	trade := GetShipType(TypeTradeShip)
	fleet.AddShip(trade, 3)

	if fleet.GetShipCount(TypeSmallShip) != 5 {
		t.Errorf("expected 5 small ships, got %d", fleet.GetShipCount(TypeSmallShip))
	}
	if fleet.GetShipCount(TypeTradeShip) != 3 {
		t.Errorf("expected 3 trade ships, got %d", fleet.GetShipCount(TypeTradeShip))
	}
	if fleet.TotalShipCount() != 8 {
		t.Errorf("expected 8 total ships, got %d", fleet.TotalShipCount())
	}
}

func TestFleetRemoveShip(t *testing.T) {
	fleet := NewFleet()
	small := GetShipType(TypeSmallShip)
	fleet.AddShip(small, 5)

	remaining := fleet.RemoveShip(TypeSmallShip, 2)
	if remaining != 3 {
		t.Errorf("expected 3 remaining ships, got %d", remaining)
	}

	remaining = fleet.RemoveShip(TypeSmallShip, 3)
	if remaining != 0 {
		t.Errorf("expected 0 remaining ships, got %d", remaining)
	}

	if _, ok := fleet.Ships[string(TypeSmallShip)]; ok {
		t.Error("expected small ship entry to be removed")
	}
}

func TestFleetTotalDamage(t *testing.T) {
	fleet := NewFleet()
	interceptor := GetShipType(TypeInterceptor)
	fleet.AddShip(interceptor, 2)

	damage := fleet.TotalDamage()
	expected := 10.0 * 2
	if damage != expected {
		t.Errorf("expected 20 max damage for 2 interceptors, got %f", damage)
	}
}

func TestShipyardMaxSlots(t *testing.T) {
	sy := NewShipyard()
	sy.Level = 3

	slots := sy.MaxSlots(2)
	expected := 10 * (2 + 3)
	if slots != expected {
		t.Errorf("expected %d max slots, got %d", expected, slots)
	}
}

func TestShipyardCanQueueShip(t *testing.T) {
	sy := NewShipyard()
	sy.Level = 4

	st := GetShipType(TypeCruiser)

	if !sy.CanQueueShip(st, 2000, 2000, 2000, 2000, 2000) {
		t.Error("expected to be able to queue cruiser with level 4 shipyard and enough resources")
	}

	sy.Level = 3
	if sy.CanQueueShip(st, 2000, 2000, 2000, 2000, 2000) {
		t.Error("expected to not queue cruiser with level 3 shipyard (needs level 4)")
	}
}

func TestShipyardQueueAndTick(t *testing.T) {
	sy := NewShipyard()
	sy.Level = 1

	st := GetShipType(TypeSmallShip)

	err := sy.QueueShip(st)
	if err != nil {
		t.Fatalf("failed to queue ship: %v", err)
	}

	if sy.GetQueuedCount() != 1 {
		t.Errorf("expected 1 ship in queue, got %d", sy.GetQueuedCount())
	}

	food := 100.0
	composite := 100.0
	mechanisms := 100.0
	reagents := 100.0
	money := 100.0
	sy.DeductCost(st, &food, &composite, &mechanisms, &reagents, &money)

	if food != 90 {
		t.Errorf("expected food 90 after deduct, got %f", food)
	}

	// Tick until the ship is completed (removed from queue)
	for sy.GetQueuedCount() > 0 {
		sy.Tick()
	}

	if sy.GetQueuedCount() != 0 {
		t.Errorf("expected 0 ships in queue after completion, got %d", sy.GetQueuedCount())
	}
}

func TestShipyardCancelShip(t *testing.T) {
	sy := NewShipyard()
	sy.Level = 1

	st := GetShipType(TypeSmallShip)

	sy.QueueShip(st)
	cancelled, _ := sy.CancelShip()

	if cancelled == nil || cancelled.TypeID != TypeSmallShip {
		t.Error("expected to cancel small ship")
	}
	if sy.GetQueuedCount() != 0 {
		t.Errorf("expected 0 ships after cancel, got %d", sy.GetQueuedCount())
	}
}

func TestShipyardCannotQueueInsufficientResources(t *testing.T) {
	sy := NewShipyard()
	sy.Level = 1

	st := GetShipType(TypeCruiser)

	// QueueShip doesn't check resources directly, but CanQueueShip does
	// The actual resource check happens in the planet's CanBuildShip
	if sy.CanQueueShip(st, 10, 10, 10, 10, 10) {
		t.Error("expected CanQueueShip to return false with insufficient resources")
	}
}

func TestGetShipTypeUnknown(t *testing.T) {
	st := GetShipType(TypeID("nonexistent"))
	if st != nil {
		t.Error("expected nil for unknown ship type")
	}
}

func TestFleetGetShipState(t *testing.T) {
	fleet := NewFleet()
	small := GetShipType(TypeSmallShip)
	fleet.AddShip(small, 2)

	state := fleet.GetShipState()
	if _, ok := state[string(TypeSmallShip)]; !ok {
		t.Error("expected small ship in fleet state")
	}

	smallState := state[string(TypeSmallShip)].(map[string]interface{})
	count := smallState["count"]
	switch v := count.(type) {
	case int:
		if v != 2 {
			t.Errorf("expected count 2 in state, got %v", v)
		}
	case float64:
		if int(v) != 2 {
			t.Errorf("expected count 2 in state, got %v", v)
		}
	default:
		t.Errorf("unexpected count type: %T", count)
	}
}
