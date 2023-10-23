package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Config struct {
	ManifestURL string `json:"manifest_url"`
	OutputDir   string `json:"output_dir"`
}

var configFilePath = "config.json"

func LoadConfig() (*Config, error) {
	var config = &Config{}
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		cfg := &Config{
			ManifestURL: "",
			OutputDir:   "",
		}
		cfg.SaveConfig()
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (cfg *Config) SaveConfig() error {
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
