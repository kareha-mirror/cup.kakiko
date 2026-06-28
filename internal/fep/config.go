package fep

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Color string `yaml:"color"`

	EscapeTimeout int `yaml:"escapce-timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Color: "252,235",

		EscapeTimeout: 100,
	}
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	return &cfg
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0666)
	if err != nil {
		return err
	}

	return nil
}
