package config

import (
	"encoding/json"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	DataDir          string   `json:"data_directory"`
	ScheduleDir      string   `json:"schedule_directory"`
	LogLevel         string   `json:"log_level"`
	ApiListenAddress string   `json:"api_listen_address"`
	ApiKeys          []string `json:"api_keys"` // Added for security
}

// Load reads a configuration file from the given path and returns a Config struct.
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}