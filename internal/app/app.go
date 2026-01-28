package app

import (
	"fmt"

	"github.com/lazydocs/lazydocs/internal/config"
	"github.com/lazydocs/lazydocs/internal/data"
	"github.com/lazydocs/lazydocs/internal/db"
	"github.com/lazydocs/lazydocs/internal/model"
)

// App is the main application orchestrator
type App struct {
	paths    config.Paths
	config   config.Config
	db       *db.DB
	client   *data.Client
	manifest *data.ManifestCache
	storage  *data.Storage
	indexer  *db.Indexer
	searcher *db.Searcher
}

// New creates a new App instance
func New() (*App, error) {
	paths := config.DefaultPaths()

	// Ensure directories exist
	if err := paths.EnsureDirs(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Load config
	cfg, err := config.Load(paths.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Open database
	database, err := db.Open(paths.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create components
	client := data.NewClient()
	manifest := data.NewManifestCache(paths.ManifestPath, client)
	storage := data.NewStorage(paths.DocsDir)
	indexer := db.NewIndexer(database)
	searcher := db.NewSearcher(database)

	return &App{
		paths:    paths,
		config:   cfg,
		db:       database,
		client:   client,
		manifest: manifest,
		storage:  storage,
		indexer:  indexer,
		searcher: searcher,
	}, nil
}

// Close cleans up app resources
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// ListInstalledDocsets returns all installed docsets
func (a *App) ListInstalledDocsets() ([]model.Docset, error) {
	return a.searcher.ListDocsets()
}

// ListAvailableDocsets returns all available docsets from DevDocs
func (a *App) ListAvailableDocsets(forceRefresh bool) (model.Manifest, error) {
	return a.manifest.Get(forceRefresh)
}

// InstallDocset downloads and indexes a docset
func (a *App) InstallDocset(slug string, progress data.ProgressCallback) error {
	// Get manifest entry
	manifest, err := a.manifest.Get(false)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	var entry *model.ManifestEntry
	for i := range manifest {
		if manifest[i].Slug == slug {
			entry = &manifest[i]
			break
		}
	}

	if entry == nil {
		return fmt.Errorf("docset %q not found in manifest", slug)
	}

	// Download and index
	downloader := data.NewDownloader(a.client, a.indexer, a.storage)
	return downloader.Download(*entry, progress)
}

// RemoveDocset removes a docset
func (a *App) RemoveDocset(slug string) error {
	// Remove from database
	if err := a.indexer.RemoveDocset(slug); err != nil {
		return fmt.Errorf("failed to remove from database: %w", err)
	}

	// Remove from disk
	if err := a.storage.DeleteDocset(slug); err != nil {
		return fmt.Errorf("failed to remove from disk: %w", err)
	}

	return nil
}

// Search performs a full-text search
func (a *App) Search(query, docset, version string, limit int) ([]model.SearchResult, error) {
	return a.searcher.Search(query, docset, version, limit)
}

// ListEntries returns entries for a docset (for browsing)
func (a *App) ListEntries(docset, version string, limit int) ([]model.Entry, error) {
	return a.searcher.ListEntries(docset, version, limit)
}

// GetEntry returns a specific entry
func (a *App) GetEntry(docset, version, path string) (*model.Entry, error) {
	return a.searcher.GetEntry(docset, version, path)
}

// Paths returns the app paths
func (a *App) Paths() config.Paths {
	return a.paths
}

// Config returns the app config
func (a *App) Config() config.Config {
	return a.config
}
