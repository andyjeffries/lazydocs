package db

import (
	"fmt"

	"github.com/lazydocs/lazydocs/internal/model"
)

// Indexer handles bulk document indexing
type Indexer struct {
	db *DB
}

// NewIndexer creates a new Indexer
func NewIndexer(db *DB) *Indexer {
	return &Indexer{db: db}
}

// IndexDocset indexes all entries for a docset
func (idx *Indexer) IndexDocset(docset model.Docset, entries []model.Entry) error {
	tx, err := idx.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, remove any existing entries for this docset
	_, err = tx.Exec(
		"DELETE FROM docs WHERE docset = ? AND version = ?",
		docset.Name, docset.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to delete existing entries: %w", err)
	}

	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO docs (docset, version, symbol, title, content, path)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert all entries
	for _, entry := range entries {
		_, err = stmt.Exec(
			entry.Docset,
			entry.Version,
			entry.Symbol,
			entry.Title,
			entry.Content,
			entry.Path,
		)
		if err != nil {
			return fmt.Errorf("failed to insert entry %s: %w", entry.Symbol, err)
		}
	}

	// Insert or update docset metadata
	_, err = tx.Exec(`
		INSERT INTO docsets (slug, name, version, display_name, entry_count, mtime)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			display_name = excluded.display_name,
			entry_count = excluded.entry_count,
			mtime = excluded.mtime,
			installed_at = strftime('%s', 'now')
	`, docset.Slug, docset.Name, docset.Version, docset.DisplayName, len(entries), docset.Mtime)
	if err != nil {
		return fmt.Errorf("failed to update docset metadata: %w", err)
	}

	return tx.Commit()
}

// RemoveDocset removes all entries and metadata for a docset
func (idx *Indexer) RemoveDocset(slug string) error {
	name, version := model.ParseSlug(slug)

	tx, err := idx.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove entries
	_, err = tx.Exec("DELETE FROM docs WHERE docset = ? AND version = ?", name, version)
	if err != nil {
		return fmt.Errorf("failed to delete entries: %w", err)
	}

	// Remove docset metadata
	_, err = tx.Exec("DELETE FROM docsets WHERE slug = ?", slug)
	if err != nil {
		return fmt.Errorf("failed to delete docset metadata: %w", err)
	}

	return tx.Commit()
}
