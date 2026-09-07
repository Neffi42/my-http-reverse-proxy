package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Listen string            `json:"listen"`
	Routes map[string]string `json:"routes"`
}

func New(path string) (*Config, error) {
	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	if len(config.Routes) == 0 {
		return nil, fmt.Errorf("at least one route is required")
	}

	return &config, nil
}
