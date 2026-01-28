package data

import (
	"os"
	"path/filepath"

	"github.com/lazydocs/lazydocs/internal/model"
)

// Storage handles local file storage for docsets
type Storage struct {
	baseDir string
}

// NewStorage creates a new Storage instance
func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

// SaveDocset saves the raw docset data to disk
func (s *Storage) SaveDocset(slug string, data []byte) error {
	name, version := model.ParseSlug(slug)

	dir := s.docsetDir(name, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "db.json")
	return os.WriteFile(path, data, 0644)
}

// LoadDocset loads the raw docset data from disk
func (s *Storage) LoadDocset(slug string) ([]byte, error) {
	name, version := model.ParseSlug(slug)
	path := filepath.Join(s.docsetDir(name, version), "db.json")
	return os.ReadFile(path)
}

// DeleteDocset removes a docset from disk
func (s *Storage) DeleteDocset(slug string) error {
	name, version := model.ParseSlug(slug)
	dir := s.docsetDir(name, version)
	return os.RemoveAll(dir)
}

// DocsetExists checks if a docset exists on disk
func (s *Storage) DocsetExists(slug string) bool {
	name, version := model.ParseSlug(slug)
	path := filepath.Join(s.docsetDir(name, version), "db.json")
	_, err := os.Stat(path)
	return err == nil
}

// docsetDir returns the directory for a docset
func (s *Storage) docsetDir(name, version string) string {
	if version != "" {
		return filepath.Join(s.baseDir, name, version)
	}
	return filepath.Join(s.baseDir, name)
}
