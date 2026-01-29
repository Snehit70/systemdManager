package service

import (
	"encoding/json"
	"testing"
)

func TestUnitJSONParsing(t *testing.T) {
	jsonData := `{
		"unit": "test.service",
		"load": "loaded",
		"active": "active",
		"sub": "running",
		"description": "Test Service"
	}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unit.Unit != "test.service" {
		t.Errorf("Expected unit 'test.service', got '%s'", unit.Unit)
	}
	if unit.Load != "loaded" {
		t.Errorf("Expected load 'loaded', got '%s'", unit.Load)
	}
	if unit.Active != "active" {
		t.Errorf("Expected active 'active', got '%s'", unit.Active)
	}
	if unit.Sub != "running" {
		t.Errorf("Expected sub 'running', got '%s'", unit.Sub)
	}
	if unit.Description != "Test Service" {
		t.Errorf("Expected description 'Test Service', got '%s'", unit.Description)
	}
}

func TestMultipleUnitsJSONParsing(t *testing.T) {
	jsonData := `[
		{
			"unit": "service1.service",
			"load": "loaded",
			"active": "active",
			"sub": "running",
			"description": "Service 1"
		},
		{
			"unit": "service2.service",
			"load": "loaded",
			"active": "failed",
			"sub": "dead",
			"description": "Service 2"
		},
		{
			"unit": "service3.service",
			"load": "loaded",
			"active": "inactive",
			"sub": "dead",
			"description": "Service 3"
		}
	]`

	var units []Unit
	err := json.Unmarshal([]byte(jsonData), &units)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(units) != 3 {
		t.Fatalf("Expected 3 units, got %d", len(units))
	}

	if units[0].Unit != "service1.service" {
		t.Errorf("Expected first unit 'service1.service', got '%s'", units[0].Unit)
	}
	if units[0].Active != "active" {
		t.Errorf("Expected first unit active 'active', got '%s'", units[0].Active)
	}

	if units[1].Unit != "service2.service" {
		t.Errorf("Expected second unit 'service2.service', got '%s'", units[1].Unit)
	}
	if units[1].Active != "failed" {
		t.Errorf("Expected second unit active 'failed', got '%s'", units[1].Active)
	}

	if units[2].Unit != "service3.service" {
		t.Errorf("Expected third unit 'service3.service', got '%s'", units[2].Unit)
	}
	if units[2].Active != "inactive" {
		t.Errorf("Expected third unit active 'inactive', got '%s'", units[2].Active)
	}
}

func TestUnitJSONParsingWithMissingFields(t *testing.T) {
	jsonData := `{
		"unit": "test.service",
		"load": "loaded"
	}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unit.Unit != "test.service" {
		t.Errorf("Expected unit 'test.service', got '%s'", unit.Unit)
	}
	if unit.Active != "" {
		t.Errorf("Expected empty active field, got '%s'", unit.Active)
	}
	if unit.Description != "" {
		t.Errorf("Expected empty description, got '%s'", unit.Description)
	}
}

func TestUnitJSONParsingInvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
