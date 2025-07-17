package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_Success tests the successful loading of a valid configuration file.
func TestLoad_Success(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// The content of our temporary config file
	configContent := `{
		"data_directory": "/tmp/data",
		"schedule_directory": "/tmp/schu",
		"log_level": "debug",
		"api_listen_address": ":9090"
	}`

	// Write the temporary config file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temporary config file: %v", err)
	}

	// Call the function we are testing
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	// Assert that the values were loaded correctly
	if cfg.DataDir != "/tmp/data" {
		t.Errorf("expected DataDir '/tmp/data', got '%s'", cfg.DataDir)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
	if cfg.ApiListenAddress != ":9090" {
		t.Errorf("expected ApiListenAddress ':9090', got '%s'", cfg.ApiListenAddress)
	}
}

// TestLoad_FileNotFound tests that Load returns an error if the file does not exist.
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("non_existent_file.json")
	if err == nil {
		t.Fatal("Load() should have returned an error for a non-existent file, but it didn't")
	}
}

// TestLoad_InvalidJSON tests that Load returns an error for a malformed JSON file.
func TestLoad_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	// Write an invalid JSON string (missing a closing brace)
	if err := os.WriteFile(configPath, []byte(`{"log_level": "info"`), 0644); err != nil {
		t.Fatalf("failed to write temporary config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() should have returned an error for invalid JSON, but it didn't")
	}
}