package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Polymarket PolymarketConfig `mapstructure:"polymarket"`
	Monitor    MonitorConfig    `mapstructure:"monitor"`
	Telegram   TelegramConfig   `mapstructure:"telegram"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// PolymarketConfig holds Polymarket API configuration
type PolymarketConfig struct {
	APIBaseURL   string        `mapstructure:"api_base_url"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
	Categories   []string      `mapstructure:"categories"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// MonitorConfig holds monitoring behavior configuration
type MonitorConfig struct {
	Threshold float64       `mapstructure:"threshold"`
	Window    time.Duration `mapstructure:"window"`
	TopK      int           `mapstructure:"top_k"`
	Enabled   bool          `mapstructure:"enabled"`
}

// TelegramConfig holds Telegram notification configuration
type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
	Enabled  bool   `mapstructure:"enabled"`
}

// StorageConfig holds storage and persistence configuration
type StorageConfig struct {
	MaxEvents            int           `mapstructure:"max_events"`
	MaxSnapshotsPerEvent int           `mapstructure:"max_snapshots_per_event"`
	MaxFileSizeMB        int           `mapstructure:"max_file_size_mb"`
	PersistenceInterval  time.Duration `mapstructure:"persistence_interval"`
	FilePath             string        `mapstructure:"file_path"`
	DataDir              string        `mapstructure:"data_dir"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file and environment variables
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigFile(path)

	// Set defaults
	setDefaults(v)

	// Enable environment variable override
	v.SetEnvPrefix("POLY_ORACLE")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures default values for all configuration options
func setDefaults(v *viper.Viper) {
	// Polymarket defaults
	v.SetDefault("polymarket.api_base_url", "https://gamma-api.polymarket.com")
	v.SetDefault("polymarket.poll_interval", "5m")
	v.SetDefault("polymarket.timeout", "30s")

	// Monitor defaults
	v.SetDefault("monitor.threshold", 0.10)
	v.SetDefault("monitor.window", "1h")
	v.SetDefault("monitor.top_k", 10)
	v.SetDefault("monitor.enabled", true)

	// Storage defaults
	v.SetDefault("storage.max_events", 1000)
	v.SetDefault("storage.max_snapshots_per_event", 100)
	v.SetDefault("storage.max_file_size_mb", 100)
	v.SetDefault("storage.persistence_interval", "5m")
	v.SetDefault("storage.file_path", "./data/poly-oracle.json")
	v.SetDefault("storage.data_dir", "./data")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// Validate checks that all configuration values are valid
func (c *Config) Validate() error {
	// Validate Polymarket config
	if c.Polymarket.APIBaseURL == "" {
		return fmt.Errorf("polymarket.api_base_url is required")
	}
	if c.Polymarket.PollInterval < 1*time.Minute {
		return fmt.Errorf("polymarket.poll_interval must be at least 1 minute")
	}
	if len(c.Polymarket.Categories) == 0 {
		return fmt.Errorf("polymarket.categories must contain at least one category")
	}

	// Validate Monitor config
	if c.Monitor.Threshold < 0.0 || c.Monitor.Threshold > 1.0 {
		return fmt.Errorf("monitor.threshold must be between 0.0 and 1.0")
	}
	if c.Monitor.Window < 1*time.Minute {
		return fmt.Errorf("monitor.window must be at least 1 minute")
	}
	if c.Monitor.TopK < 1 {
		return fmt.Errorf("monitor.top_k must be at least 1")
	}

	// Validate Telegram config
	if c.Telegram.Enabled {
		if c.Telegram.BotToken == "" {
			return fmt.Errorf("telegram.bot_token is required when telegram is enabled")
		}
		if c.Telegram.ChatID == "" {
			return fmt.Errorf("telegram.chat_id is required when telegram is enabled")
		}
	}

	// Validate Storage config
	if c.Storage.MaxEvents < 1 {
		return fmt.Errorf("storage.max_events must be at least 1")
	}
	if c.Storage.MaxSnapshotsPerEvent < 10 {
		return fmt.Errorf("storage.max_snapshots_per_event must be at least 10")
	}
	if c.Storage.MaxFileSizeMB < 1 {
		return fmt.Errorf("storage.max_file_size_mb must be at least 1")
	}
	if c.Storage.PersistenceInterval < 1*time.Minute {
		return fmt.Errorf("storage.persistence_interval must be at least 1 minute")
	}
	if c.Storage.FilePath == "" {
		return fmt.Errorf("storage.file_path is required")
	}

	// Validate Logging config
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("logging.format must be one of: json, text")
	}

	return nil
}

// GetPolymarketConfig returns the Polymarket configuration
func (c *Config) GetPolymarketConfig() PolymarketConfig {
	return c.Polymarket
}

// GetMonitorConfig returns the Monitor configuration
func (c *Config) GetMonitorConfig() MonitorConfig {
	return c.Monitor
}

// GetTelegramConfig returns the Telegram configuration
func (c *Config) GetTelegramConfig() TelegramConfig {
	return c.Telegram
}

// GetStorageConfig returns the Storage configuration
func (c *Config) GetStorageConfig() StorageConfig {
	return c.Storage
}

// GetLoggingConfig returns the Logging configuration
func (c *Config) GetLoggingConfig() LoggingConfig {
	return c.Logging
}
