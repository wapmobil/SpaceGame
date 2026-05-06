package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStartExpedition(t *testing.T) {
	handler := handleStartExpedition(nil)
	if handler == nil {
		t.Fatal("expected handleStartExpedition to be non-nil")
	}
}

func TestHandleStartExpedition_InvalidBody(t *testing.T) {
	handler := handleStartExpedition(nil)
	if handler == nil {
		t.Fatal("expected handleStartExpedition to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/expeditions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleGetExpeditionChains(t *testing.T) {
	handler := handleGetExpeditionChains(nil)
	if handler == nil {
		t.Fatal("expected handleGetExpeditionChains to be non-nil")
	}
}

func TestHandleGetExpeditionEvent(t *testing.T) {
	handler := handleGetExpeditionEvent(nil)
	if handler == nil {
		t.Fatal("expected handleGetExpeditionEvent to be non-nil")
	}
}

func TestHandleExpeditionChoice(t *testing.T) {
	handler := handleExpeditionChoice(nil)
	if handler == nil {
		t.Fatal("expected handleExpeditionChoice to be non-nil")
	}
}

func TestHandleExpeditionChoice_InvalidBody(t *testing.T) {
	handler := handleExpeditionChoice(nil)
	if handler == nil {
		t.Fatal("expected handleExpeditionChoice to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/expeditions/chain1/choice", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleGetExpeditionEvents(t *testing.T) {
	handler := handleGetExpeditionEvents(nil)
	if handler == nil {
		t.Fatal("expected handleGetExpeditionEvents to be non-nil")
	}
}

func TestHandleGetExpeditionEventLog(t *testing.T) {
	handler := handleGetExpeditionEventLog(nil)
	if handler == nil {
		t.Fatal("expected handleGetExpeditionEventLog to be non-nil")
	}
}

func TestHandleStartExpedition_InvalidInventory(t *testing.T) {
	handler := handleStartExpedition(nil)
	if handler == nil {
		t.Fatal("expected handleStartExpedition to be non-nil")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"inventory": map[string]interface{}{
			"food": "not_a_number",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/planets/test-expeditions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid inventory type, got %d: %s", w.Code, w.Body.String())
	}
}



func TestHandleGetLocations(t *testing.T) {
	handler := handleGetLocations(nil)
	if handler == nil {
		t.Fatal("expected handleGetLocations to be non-nil")
	}
}

func TestHandleBuildOnLocation(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}
}

func TestHandleBuildOnLocation_InvalidBody(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/locations/some-id/build", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleBuildOnLocation_MissingBuildingType(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/locations/some-id/build", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 status for missing building_type, got %d", w.Code)
	}
}

func TestHandleRemoveBuilding(t *testing.T) {
	handler := handleRemoveBuilding(nil)
	if handler == nil {
		t.Fatal("expected handleRemoveBuilding to be non-nil")
	}
}

func TestHandleAbandonLocation(t *testing.T) {
	handler := handleAbandonLocation(nil)
	if handler == nil {
		t.Fatal("expected handleAbandonLocation to be non-nil")
	}
}

func TestGetLocationBuildingTypes(t *testing.T) {
	types := getLocationBuildingTypes("pond")
	if len(types) != 2 {
		t.Errorf("expected 2 building types for pond, got %d", len(types))
	}

	foundFishFarm := false
	foundWaterPurifier := false
	for _, bt := range types {
		if bt == "fish_farm" {
			foundFishFarm = true
		}
		if bt == "water_purifier" {
			foundWaterPurifier = true
		}
	}
	if !foundFishFarm {
		t.Error("expected fish_farm in pond building types")
	}
	if !foundWaterPurifier {
		t.Error("expected water_purifier in pond building types")
	}
}

func TestGetLocationBuildingTypes_Unknown(t *testing.T) {
	types := getLocationBuildingTypes("nonexistent_type")
	if types != nil {
		t.Errorf("expected nil for unknown location type, got %v", types)
	}
}

func TestHandleBuildOnLocation_InvalidPlanetID(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/invalid-uuid-format/locations/some-id/build", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 status for invalid planet, got %d", w.Code)
	}
}

func TestHandleBuildOnLocation_EmptyBody(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/locations/some-id/build", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty body, got %d", w.Code)
	}
}
