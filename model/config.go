package model

import (
	"encoding/json"
	"fmt"
	"os"
)

type ClientConfig struct {
	PrivateKeyPath string `json:"private_key_path"`
	PublicKeyPath  string `json:"public_key_path"`
}

// Config defines the overall structure of the config.json file.
type Config struct {
	Server   map[string]interface{}  `json:"server"`
	Database map[string]interface{}  `json:"database"`
	Clients  map[string]ClientConfig `json:"clients"`
	Helper   map[string]interface{}  `json:"helper"`
}

func LoadConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}
	return &config, nil
}
