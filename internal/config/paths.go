package config

import (
	"os"
	"path/filepath"
)

// Paths holds all the filesystem paths used by lazydocs
type Paths struct {
	DataDir     string // Base data directory
	DocsDir     string // Where docsets are stored
	DBPath      string // SQLite database path
	ManifestPath string // Cached manifest path
	ConfigPath  string // User configuration path
}

// DefaultPaths returns the default paths following XDG conventions
func DefaultPaths() Paths {
	dataDir := getDataDir()
	configDir := getConfigDir()

	return Paths{
		DataDir:      dataDir,
		DocsDir:      filepath.Join(dataDir, "docs"),
		DBPath:       filepath.Join(dataDir, "index.sqlite"),
		ManifestPath: filepath.Join(dataDir, "manifest.json"),
		ConfigPath:   filepath.Join(configDir, "config.yaml"),
	}
}

// EnsureDirs creates all necessary directories
func (p Paths) EnsureDirs() error {
	dirs := []string{
		p.DataDir,
		p.DocsDir,
		filepath.Dir(p.ConfigPath),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// getDataDir returns the data directory following XDG conventions
func getDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "lazydocs")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".lazydocs"
	}

	return filepath.Join(home, ".local", "share", "lazydocs")
}

// getConfigDir returns the config directory following XDG conventions
func getConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "lazydocs")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".lazydocs"
	}

	return filepath.Join(home, ".config", "lazydocs")
}
