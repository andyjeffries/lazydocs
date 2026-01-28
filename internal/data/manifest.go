package data

import (
	"encoding/json"
	"os"
	"time"

	"github.com/lazydocs/lazydocs/internal/model"
)

// CachedManifest wraps a manifest with cache metadata
type CachedManifest struct {
	Manifest  model.Manifest `json:"manifest"`
	FetchedAt time.Time      `json:"fetched_at"`
}

// ManifestCache handles manifest caching
type ManifestCache struct {
	path     string
	maxAge   time.Duration
	client   *Client
	manifest model.Manifest
}

// NewManifestCache creates a new manifest cache
func NewManifestCache(path string, client *Client) *ManifestCache {
	return &ManifestCache{
		path:   path,
		maxAge: 24 * time.Hour, // Cache for 24 hours
		client: client,
	}
}

// Get returns the manifest, using cache if valid
func (mc *ManifestCache) Get(forceRefresh bool) (model.Manifest, error) {
	if !forceRefresh {
		cached, err := mc.loadFromDisk()
		if err == nil && time.Since(cached.FetchedAt) < mc.maxAge {
			mc.manifest = cached.Manifest
			return cached.Manifest, nil
		}
	}

	// Fetch fresh manifest
	manifest, err := mc.client.FetchManifest()
	if err != nil {
		// If we have a cached version, use it even if expired
		if mc.manifest != nil {
			return mc.manifest, nil
		}
		return nil, err
	}

	mc.manifest = manifest

	// Save to disk (ignore errors, cache is best-effort)
	_ = mc.saveToDisk(manifest)

	return manifest, nil
}

// Find finds a docset by slug in the manifest
func (mc *ManifestCache) Find(slug string) *model.ManifestEntry {
	for i := range mc.manifest {
		if mc.manifest[i].Slug == slug {
			return &mc.manifest[i]
		}
	}
	return nil
}

// loadFromDisk loads the cached manifest from disk
func (mc *ManifestCache) loadFromDisk() (*CachedManifest, error) {
	data, err := os.ReadFile(mc.path)
	if err != nil {
		return nil, err
	}

	var cached CachedManifest
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	return &cached, nil
}

// saveToDisk saves the manifest to disk
func (mc *ManifestCache) saveToDisk(manifest model.Manifest) error {
	cached := CachedManifest{
		Manifest:  manifest,
		FetchedAt: time.Now(),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return os.WriteFile(mc.path, data, 0644)
}
