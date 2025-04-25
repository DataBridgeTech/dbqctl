package internal

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

type ConfigDetails struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database,omitempty"`
}
