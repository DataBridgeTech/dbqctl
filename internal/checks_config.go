package internal

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type ChecksConfig struct {
	Version     string       `yaml:"version"`
	Validations []Validation `yaml:"validations"`
}

type ConfigDetails struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database,omitempty"`
}

type Validation struct {
	Dataset string  `yaml:"dataset"`
	Checks  []Check `yaml:"checks"`
}

type Check struct {
	ID          string                 `yaml:"id"`
	Description string                 `yaml:"description,omitempty"`
	Severity    string                 `yaml:"severity"`
	Type        string                 `yaml:"type"`
	Params      map[string]interface{} `yaml:"params"`
}

func LoadChecksConfig(fileName string) (*ChecksConfig, error) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return nil, err
	}

	var cfg ChecksConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		log.Printf("Error decoding YAML: %v\n", err)
		return nil, err
	}

	return &cfg, nil
}
