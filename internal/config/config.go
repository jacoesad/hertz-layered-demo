package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Logger     LoggerConfig     `yaml:"logger"`
	Database   DatabaseConfig   `yaml:"database"`
	Signature  SignatureConfig  `yaml:"signature"`
	Downstream DownstreamConfig `yaml:"downstream"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type LoggerConfig struct {
	Level     string `yaml:"level"`
	FilePath  string `yaml:"file_path"`
	MaxSizeMB int    `yaml:"max_size_mb"`
}

type DatabaseConfig struct {
	Main      DataSourceConfig `yaml:"main"`
	StarRocks DataSourceConfig `yaml:"starrocks"`
}

type DataSourceConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type SignatureConfig struct {
	Enabled bool   `yaml:"enabled"`
	Secret  string `yaml:"secret"`
	Header  string `yaml:"header"`
}

type DownstreamConfig struct {
	TaskRunner TaskRunnerConfig `yaml:"task_runner"`
}

type TaskRunnerConfig struct {
	Endpoint  string `yaml:"endpoint"`
	TimeoutMS int    `yaml:"timeout_ms"`
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
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.FilePath == "" {
		return nil, fmt.Errorf("logger.file_path is required")
	}
	if cfg.Logger.MaxSizeMB <= 0 {
		return nil, fmt.Errorf("logger.max_size_mb must be positive")
	}
	if cfg.Database.Main.Driver == "" {
		return nil, fmt.Errorf("database.main.driver is required")
	}
	if cfg.Database.Main.DSN == "" {
		return nil, fmt.Errorf("database.main.dsn is required")
	}
	if cfg.Database.StarRocks.Driver == "" {
		return nil, fmt.Errorf("database.starrocks.driver is required")
	}
	if cfg.Database.StarRocks.DSN == "" {
		return nil, fmt.Errorf("database.starrocks.dsn is required")
	}
	if cfg.Signature.Enabled && cfg.Signature.Secret == "" {
		return nil, fmt.Errorf("signature.secret is required when signature is enabled")
	}
	if cfg.Signature.Header == "" {
		cfg.Signature.Header = "X-Signature"
	}
	if cfg.Downstream.TaskRunner.Endpoint == "" {
		return nil, fmt.Errorf("downstream.task_runner.endpoint is required")
	}
	if cfg.Downstream.TaskRunner.TimeoutMS <= 0 {
		return nil, fmt.Errorf("downstream.task_runner.timeout_ms must be positive")
	}
	return &cfg, nil
}
