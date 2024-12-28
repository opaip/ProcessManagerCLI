package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// SaveToFile saves data to a specified file
func SaveToFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

func LoadFromFile(filename string, target interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// ReadConfig reads a configuration file and returns its contents
func ReadConfig(filename string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config map[string]interface{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return config, nil
}

func (p *Process) SaveStat() error {
	// Get the current date in YYYY-MM-DD format
	currentDate := time.Now().Format("2006-01-02")

	// Create a folder with the current date if it doesn't exist
	err := os.MkdirAll("./data/"+currentDate, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Marshal the process data into JSON
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal process data: %v", err)
	}

	// Save the JSON data to a file inside the folder named by the current date
	err = ioutil.WriteFile("./data/"+currentDate+"/"+p.Name+"_status.json", data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save process status: %v", err)
	}

	return nil
}
