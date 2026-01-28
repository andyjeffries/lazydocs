package data

import (
	"encoding/json"
	"fmt"

	"github.com/lazydocs/lazydocs/internal/db"
	"github.com/lazydocs/lazydocs/internal/model"
)

// ProgressCallback is called during download with progress updates
type ProgressCallback func(downloaded, total int64, status string)

// Downloader handles downloading and indexing docsets
type Downloader struct {
	client    *Client
	converter *Converter
	indexer   *db.Indexer
	storage   *Storage
}

// NewDownloader creates a new Downloader
func NewDownloader(client *Client, indexer *db.Indexer, storage *Storage) *Downloader {
	return &Downloader{
		client:    client,
		converter: NewConverter(),
		indexer:   indexer,
		storage:   storage,
	}
}

// Download downloads, converts, and indexes a docset
func (d *Downloader) Download(entry model.ManifestEntry, progress ProgressCallback) error {
	if progress != nil {
		progress(0, entry.DBSize, "Downloading...")
	}

	// Download the docset data
	data, err := d.client.FetchDocset(entry.Slug, func(downloaded, total int64) {
		if progress != nil {
			progress(downloaded, total, "Downloading...")
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download docset: %w", err)
	}

	if progress != nil {
		progress(entry.DBSize, entry.DBSize, "Processing...")
	}

	// Save raw data
	if err := d.storage.SaveDocset(entry.Slug, data); err != nil {
		return fmt.Errorf("failed to save docset: %w", err)
	}

	// Fetch the index to get entry metadata
	indexData, err := d.client.FetchIndex(entry.Slug)
	if err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}

	// Parse and convert the content
	entries, err := d.parseDocset(entry, data, indexData)
	if err != nil {
		return fmt.Errorf("failed to parse docset: %w", err)
	}

	if progress != nil {
		progress(entry.DBSize, entry.DBSize, "Indexing...")
	}

	// Index into database
	name, version := model.ParseSlug(entry.Slug)
	docset := model.Docset{
		Slug:        entry.Slug,
		Name:        name,
		Version:     version,
		DisplayName: entry.Name,
		EntryCount:  len(entries),
		Mtime:       entry.Mtime,
	}

	if err := d.indexer.IndexDocset(docset, entries); err != nil {
		return fmt.Errorf("failed to index docset: %w", err)
	}

	if progress != nil {
		progress(entry.DBSize, entry.DBSize, "Done")
	}

	return nil
}

// parseDocset parses the raw docset JSON and converts HTML to markdown
func (d *Downloader) parseDocset(manifest model.ManifestEntry, data []byte, indexData *DocsetData) ([]model.Entry, error) {
	// DevDocs db.json format is a map of path -> HTML content
	var contentMap map[string]string
	if err := json.Unmarshal(data, &contentMap); err != nil {
		return nil, fmt.Errorf("failed to parse docset JSON: %w", err)
	}

	name, version := model.ParseSlug(manifest.Slug)

	// Build a map for quick lookup of entry metadata
	entryMeta := make(map[string]DocsetEntry)
	for _, e := range indexData.Entries {
		// Path may contain fragment, normalize it
		path := e.Path
		if idx := indexOf(path, '#'); idx != -1 {
			path = path[:idx]
		}
		entryMeta[path] = e
	}

	var entries []model.Entry
	for path, html := range contentMap {
		// Convert HTML to Markdown
		markdown, err := d.converter.Convert(html)
		if err != nil {
			// Skip entries that fail to convert
			continue
		}

		// Get metadata from index
		meta, ok := entryMeta[path]
		symbol := path
		title := path
		if ok {
			symbol = meta.Name
			title = meta.Name
		}

		entries = append(entries, model.Entry{
			Docset:  name,
			Version: version,
			Symbol:  symbol,
			Title:   title,
			Content: markdown,
			Path:    path,
		})
	}

	return entries, nil
}

// indexOf returns the index of c in s, or -1 if not found
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
