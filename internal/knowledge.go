package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// dataDir returns the path to the data/ directory relative to the project root.
func dataDir() string {
	// Try relative to the binary's working directory first
	if _, err := os.Stat("data"); err == nil {
		return "data"
	}
	// Fallback: relative to this source file (for tests)
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "data")
}

func LoadSystemDescription() (*SystemDescription, error) {
	data, err := os.ReadFile(filepath.Join(dataDir(), "system_description.json"))
	if err != nil {
		return nil, fmt.Errorf("load system description: %w", err)
	}
	var sd SystemDescription
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("parse system description: %w", err)
	}
	return &sd, nil
}

func LoadPastIncidents() ([]PastIncident, error) {
	data, err := os.ReadFile(filepath.Join(dataDir(), "past_incidents.json"))
	if err != nil {
		return nil, fmt.Errorf("load past incidents: %w", err)
	}
	var incidents []PastIncident
	if err := json.Unmarshal(data, &incidents); err != nil {
		return nil, fmt.Errorf("parse past incidents: %w", err)
	}
	return incidents, nil
}
