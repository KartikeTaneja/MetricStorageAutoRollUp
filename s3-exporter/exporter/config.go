package exporter

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds all configuration for the S3 exporter
type Config struct {
	S3 struct {
		Region    string `yaml:"region"`
		Bucket    string `yaml:"bucket"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"s3"`

	Export struct {
		BatchSize   int    `yaml:"batch_size"`
		Compression bool   `yaml:"compression"`
		TempDir     string `yaml:"temp_dir"`
	} `yaml:"export"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Create default config
	config := &Config{}
	
	// Set defaults
	config.Export.BatchSize = 1000
	config.Export.Compression = true
	config.Export.TempDir = "/tmp/s3-exporter"
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	
	return config, nil
}