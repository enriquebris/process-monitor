package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Entries []MonitorEntry `json:"processes"`
}

// loadConfig loads && returns Config from a given filepath
func loadConfig(filepath string) (*Config, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	return &config, err
}
