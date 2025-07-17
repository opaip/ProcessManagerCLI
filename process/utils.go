package process

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveToFile marshals data to JSON and saves it to a file.
func SaveToFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	return encoder.Encode(data)
}

// LoadFromFile reads a JSON file and unmarshals it into the target interface.
func LoadFromFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode json from %s: %w", filePath, err)
	}
	return nil
}

// FileExists checks if a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}