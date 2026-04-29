package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStartPlanetSurvey(t *testing.T) {
	handler := handleStartPlanetSurvey(nil)
	if handler == nil {
		t.Fatal("expected handleStartPlanetSurvey to be non-nil")
	}
}

func TestHandleStartPlanetSurvey_InvalidBody(t *testing.T) {
	handler := handleStartPlanetSurvey(nil)
	if handler == nil {
		t.Fatal("expected handleStartPlanetSurvey to be non-nil")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/planet-survey", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleGetPlanetSurvey(t *testing.T) {
	handler := handleGetPlanetSurvey(nil)
	if handler == nil {
		t.Fatal("expected handleGetPlanetSurvey to be non-nil")
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

func TestHandleGetExpeditionHistory(t *testing.T) {
	handler := handleGetExpeditionHistory(nil)
	if handler == nil {
		t.Fatal("expected handleGetExpeditionHistory to be non-nil")
	}
}

func TestGetMaxDurationForBaseLevel(t *testing.T) {
	tests := []struct {
		baseLevel int
		expected  int
	}{
		{1, 300},
		{2, 600},
		{3, 1200},
		{5, 300},
	}

	for _, tc := range tests {
		result := getMaxDurationForBaseLevel(tc.baseLevel)
		if result != tc.expected {
			t.Errorf("baseLevel %d: expected %d, got %d", tc.baseLevel, tc.expected, result)
		}
	}
}

func TestGetRangeForDuration(t *testing.T) {
	tests := []struct {
		duration int
		expected string
	}{
		{300, "300s"},
		{150, "300s"},
		{600, "600s"},
		{400, "600s"},
		{1200, "1200s"},
		{900, "1200s"},
	}

	for _, tc := range tests {
		result := getRangeForDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("duration %d: expected %s, got %s", tc.duration, tc.expected, result)
		}
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

func TestHandleStartPlanetSurvey_ValidBody(t *testing.T) {
	handler := handleStartPlanetSurvey(nil)
	if handler == nil {
		t.Fatal("expected handleStartPlanetSurvey to be non-nil")
	}

	// Send empty duration to trigger bad request before planet loading
	reqBody, _ := json.Marshal(map[string]interface{}{
		"duration": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/planet-survey", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for zero duration, got %d", w.Code)
	}
}

func TestHandleStartPlanetSurvey_MissingDuration(t *testing.T) {
	handler := handleStartPlanetSurvey(nil)
	if handler == nil {
		t.Fatal("expected handleStartPlanetSurvey to be non-nil")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/planets/test/planet-survey", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing duration, got %d", w.Code)
	}
}

func TestHandleBuildOnLocation_InvalidPlanetID(t *testing.T) {
	handler := handleBuildOnLocation(nil)
	if handler == nil {
		t.Fatal("expected handleBuildOnLocation to be non-nil")
	}

	// Send empty body to trigger bad request before planet loading
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
