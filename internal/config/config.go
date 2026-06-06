package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig   `yaml:"server"`
	Database  DatabaseConfig `yaml:"database"`
	StarRocks DatabaseConfig `yaml:"starrocks"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

var current *Config

func Init(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	current = cfg
	return cfg, nil
}

func Get() *Config {
	return current
}

func MustGet() *Config {
	if current == nil {
		panic("config is not initialized")
	}
	return current
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.Addr == "" {
		return nil, fmt.Errorf("server.addr is required")
	}
	if cfg.Database.Driver == "" {
		return nil, fmt.Errorf("database.driver is required")
	}
	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.dsn is required")
	}
	if cfg.StarRocks.Driver == "" {
		return nil, fmt.Errorf("starrocks.driver is required")
	}
	if cfg.StarRocks.DSN == "" {
		return nil, fmt.Errorf("starrocks.dsn is required")
	}
	return &cfg, nil
}
