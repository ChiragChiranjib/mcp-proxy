// Package config loads application configuration via Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ServerConfig holds HTTP server options.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	AESKey string `mapstructure:"aes_key"`
}

// DatabaseConfig holds MySQL connection and pool settings.
type DatabaseConfig struct {
	DSN                    string `mapstructure:"dsn"`
	Username               string `mapstructure:"username"`
	Password               string `mapstructure:"password"`
	Host                   string `mapstructure:"host"`
	Port                   int    `mapstructure:"port"`
	Name                   string `mapstructure:"name"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxIdleSeconds     int    `mapstructure:"conn_max_idle_seconds"`
	ConnMaxLifetimeSeconds int    `mapstructure:"conn_max_lifetime_seconds"`
}

// Config is the root application configuration.
type Config struct {
	AppEnv   string
	MCPMode  string
	Server   ServerConfig   `mapstructure:"server"`
	DB       DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
}

// Load reads the TOML config for the current APP_ENV and MCP_MODE.
func Load() (*Config, error) {
	appEnv := getenvDefault("APP_ENV", "dev")
	mcpMode := getenvDefault("MCP_MODE", "streamable-http")
	file := fmt.Sprintf("%s_%s.toml", appEnv, mcpMode)
	vp := viper.New()
	vp.SetConfigType("toml")
	vp.AddConfigPath("./config")
	vp.SetConfigFile(filepath.Join("./config", file))
	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", file, err)
	}
	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.AppEnv = appEnv
	cfg.MCPMode = mcpMode
	return &cfg, nil
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
