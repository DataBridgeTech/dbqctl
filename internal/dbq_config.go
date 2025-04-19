package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type DbqConfig struct {
	Version     string       `yaml:"version"`
	DataSources []DataSource `yaml:"datasources"`
}

type DataSource struct {
	ID            string        `yaml:"id"`
	Type          string        `yaml:"type"`
	Configuration ConfigDetails `yaml:"configuration"`
	Datasets      []string      `yaml:"datasets"`
}

func LoadDbqSetting(fileName string) (*DbqConfig, error) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil, err
	}

	var settings DbqConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		fmt.Printf("Error decoding YAML: %v\n", err)
		return nil, err
	}

	return &settings, nil
}
