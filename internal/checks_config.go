package internal

import (
	"gopkg.in/yaml.v3"
	"os"
)

type ChecksConfig struct {
	Version     string       `yaml:"version"`
	Validations []Validation `yaml:"validations"`
}

type Validation struct {
	Dataset string  `yaml:"dataset"`
	Where   string  `yaml:"where,omitempty"` // Optional where clause
	Checks  []Check `yaml:"checks"`
}

type Check struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description,omitempty"` // Optional
	Severity    string `yaml:"severity,omitempty"`    // Optional (error, warn, info)
	Query       string `yaml:"query,omitempty"`       // Optional raw query
}

func LoadChecksConfig(fileName string) (*ChecksConfig, error) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	var cfg ChecksConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
