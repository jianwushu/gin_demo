package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
	Database DatabaseConfig `yaml:"database"`
	Swagger  SwaggerConfig  `yaml:"swagger"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	SSL          SSLConfig     `yaml:"ssl"`
}

type SSLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type LogConfig struct {
	Level         string `yaml:"level"`
	Dir           string `yaml:"dir"`
	Filename      string `yaml:"filename"`
	RotateBy      string `yaml:"rotate_by"`
	RetentionDays int    `yaml:"retention_days"`
}

type DatabaseConfig struct {
	Driver      string `yaml:"driver"`
	DSN         string `yaml:"dsn"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

type SwaggerConfig struct {
	Enabled bool `yaml:"enabled"`
}

func Default() *Config {
	return &Config{
		App: AppConfig{
			Name: "gin-demo",
			Env:  "dev",
		},
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Log: LogConfig{
			Level:         "info",
			Dir:           "logs",
			Filename:      "app",
			RotateBy:      "day",
			RetentionDays: 7,
		},
		Database: DatabaseConfig{
			Driver:      "sqlite",
			DSN:         "data/gin-demo.db",
			AutoMigrate: true,
		},
		Swagger: SwaggerConfig{
			Enabled: false,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be greater than 0")
	}
	if c.Log.Dir == "" {
		return fmt.Errorf("log.dir is required")
	}
	if c.Log.Filename == "" {
		return fmt.Errorf("log.filename is required")
	}
	if c.Log.RotateBy != "day" && c.Log.RotateBy != "hour" {
		return fmt.Errorf("log.rotate_by must be day or hour")
	}
	if c.Log.RetentionDays <= 0 {
		return fmt.Errorf("log.retention_days must be greater than 0")
	}
	if c.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if c.Server.SSL.Enabled {
		if c.Server.SSL.CertFile == "" || c.Server.SSL.KeyFile == "" {
			return fmt.Errorf("server.ssl.cert_file and server.ssl.key_file are required when ssl is enabled")
		}
	}
	if c.Swagger.Enabled && c.App.Env != "dev" {
		return fmt.Errorf("swagger can only be enabled when app.env is dev")
	}

	return nil
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
