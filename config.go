package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var configFilePath = "config.json"

func loadConfig(config *Config) {
	data, err := ioutil.ReadFile(configFilePath)
	if err == nil {
		_ = json.Unmarshal(data, config)
	}
}

func saveConfig(config *Config) error {
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	data, _ := json.MarshalIndent(config, "", "  ")
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
