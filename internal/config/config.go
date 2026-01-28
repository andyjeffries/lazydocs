package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds user configuration
type Config struct {
	// Theme for markdown rendering: "dark", "light", "dracula", "notty"
	Theme string `yaml:"theme"`

	// DefaultDocsets are automatically selected on startup
	DefaultDocsets []string `yaml:"default_docsets,omitempty"`

	// UI customization
	UI UIConfig `yaml:"ui"`
}

// UIConfig holds UI-related settings
type UIConfig struct {
	// Show debug info in status bar
	ShowDebug bool `yaml:"show_debug"`

	// Colors (hex or named)
	PrimaryColor   string `yaml:"primary_color,omitempty"`
	SecondaryColor string `yaml:"secondary_color,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme:          "dark",
		DefaultDocsets: []string{},
		UI: UIConfig{
			ShowDebug:      false,
			PrimaryColor:   "",
			SecondaryColor: "",
		},
	}
}

// Load loads the configuration from the given path
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save saves the configuration to the given path
func (c Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CreateDefault creates a default config file if it doesn't exist
func CreateDefault(path string) error {
	if _, err := os.Stat(path); err == nil {
		// File exists, don't overwrite
		return nil
	}

	cfg := DefaultConfig()
	return cfg.Save(path)
}
