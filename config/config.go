package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration values
type Config struct {
	Audio    AudioConfig    `yaml:"audio"`
	Server   ServerConfig   `yaml:"server"`
	Workers  WorkersConfig  `yaml:"workers"`
	Retry    RetryConfig    `yaml:"retry"`
	Shutdown ShutdownConfig `yaml:"shutdown"`
	Log      LogConfig      `yaml:"log"`
}

// AudioConfig holds audio settings
type AudioConfig struct {
	SampleRate      int    `yaml:"sample_rate"`
	Channels        int    `yaml:"channels"`
	SegmentDuration int    `yaml:"segment_duration"`
	TestFile        string `yaml:"test_file"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	URL               string        `yaml:"url"`
	Timeout           time.Duration `yaml:"timeout"`
	Username          string        `yaml:"username"`
	Password          string        `yaml:"password"`
	TranscribeTimeout time.Duration `yaml:"transcribe_timeout"`
}

// WorkersConfig holds worker settings
type WorkersConfig struct {
	Count     int `yaml:"count"`
	QueueSize int `yaml:"queue_size"`
}

// RetryConfig holds retry settings
type RetryConfig struct {
	MaxRetries   int `yaml:"max_retries"`
	InitialDelay int `yaml:"initial_delay"`
	MaxDelay     int `yaml:"max_delay"`
}

// ShutdownConfig holds shutdown settings
type ShutdownConfig struct {
	Timeout int `yaml:"timeout"`
}

// LogConfig holds logging settings
type LogConfig struct {
	Level string `yaml:"level"`
}

// NewConfig returns the global configuration
func NewConfig() *Config {
	return &Config{}
}

// LoadConfig loads configuration from a YAML file
func (cfg *Config) LoadConfig(filename string) error {
	// Initialize defaults first
	cfg.initDefault()

	// Check if file exists
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file %s not found, using defaults\n", filename)
			return nil
		}
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// initDefault initializes default configuration
func (cfg *Config) initDefault() {
	cfg.Audio = AudioConfig{
		SampleRate:      16000,
		Channels:        1,
		SegmentDuration: 60,
	}
	cfg.Server = ServerConfig{
		URL:     "http://localhost:8080/transcribe",
		Timeout: time.Duration(30) * time.Second,
	}
	cfg.Workers = WorkersConfig{
		Count:     2,
		QueueSize: 10,
	}
	cfg.Retry = RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1,
		MaxDelay:     60,
	}
	cfg.Shutdown = ShutdownConfig{
		Timeout: 30,
	}
	cfg.Log = LogConfig{
		Level: "info",
	}
}

// Validate validates the configuration
func (cfg *Config) Validate() error {
	if cfg.Audio.SampleRate <= 0 {
		return errors.New("audio.sample_rate must be positive")
	}
	if cfg.Audio.Channels <= 0 {
		return errors.New("audio.channels must be positive")
	}
	if cfg.Audio.SegmentDuration < 10 || cfg.Audio.SegmentDuration > 600 {
		return errors.New("audio.segment_duration must be between 10 and 600 seconds")
	}
	if cfg.Server.URL == "" {
		return errors.New("server.url is required")
	}
	if cfg.Server.Timeout <= 0 {
		return errors.New("server.timeout must be positive")
	}
	if cfg.Workers.Count <= 0 {
		return errors.New("workers.count must be positive")
	}
	if cfg.Retry.MaxRetries < 0 {
		return errors.New("retry.max_retries must be non-negative")
	}
	if cfg.Retry.InitialDelay <= 0 {
		return errors.New("retry.initial_delay must be positive")
	}
	if cfg.Retry.MaxDelay < cfg.Retry.InitialDelay {
		return errors.New("retry.max_delay must be >= retry.initial_delay")
	}
	if cfg.Shutdown.Timeout <= 0 {
		return errors.New("shutdown.timeout must be positive")
	}

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[cfg.Log.Level] {
		return errors.New("log.level must be one of: debug, info, warn, error")
	}

	return nil
}
