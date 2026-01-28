package db

import (
	"fmt"
	"strings"

	"github.com/lazydocs/lazydocs/internal/model"
)

// Searcher handles FTS5 search queries
type Searcher struct {
	db *DB
}

// NewSearcher creates a new Searcher
func NewSearcher(db *DB) *Searcher {
	return &Searcher{db: db}
}

// Search performs a full-text search across all docsets or a specific one
func (s *Searcher) Search(query string, docset string, version string, limit int) ([]model.SearchResult, error) {
	if query == "" {
		return nil, nil
	}

	if limit <= 0 {
		limit = 50
	}

	// Build the FTS5 query
	// Escape special FTS5 characters and add prefix matching
	ftsQuery := escapeFTS5Query(query)

	var args []any
	var whereClause string

	if docset != "" {
		if version != "" {
			whereClause = "AND docset = ? AND version = ?"
			args = append(args, ftsQuery, docset, version, limit)
		} else {
			whereClause = "AND docset = ?"
			args = append(args, ftsQuery, docset, limit)
		}
	} else {
		args = append(args, ftsQuery, limit)
	}

	sql := fmt.Sprintf(`
		SELECT
			docset,
			version,
			symbol,
			title,
			content,
			path,
			bm25(docs) as rank,
			snippet(docs, 4, '<mark>', '</mark>', '...', 32) as snippet
		FROM docs
		WHERE docs MATCH ?
		%s
		ORDER BY rank
		LIMIT ?
	`, whereClause)

	rows, err := s.db.conn.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []model.SearchResult
	for rows.Next() {
		var r model.SearchResult
		err := rows.Scan(
			&r.Docset,
			&r.Version,
			&r.Symbol,
			&r.Title,
			&r.Content,
			&r.Path,
			&r.Rank,
			&r.Snippet,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// ListEntries returns all entries for a docset (for browsing without search)
func (s *Searcher) ListEntries(docset string, version string, limit int) ([]model.Entry, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows interface {
		Next() bool
		Scan(dest ...any) error
		Err() error
		Close() error
	}
	var err error

	if version != "" {
		rows, err = s.db.conn.Query(`
			SELECT docset, version, symbol, title, content, path
			FROM docs
			WHERE docset = ? AND version = ?
			ORDER BY symbol
			LIMIT ?
		`, docset, version, limit)
	} else {
		rows, err = s.db.conn.Query(`
			SELECT docset, version, symbol, title, content, path
			FROM docs
			WHERE docset = ?
			ORDER BY symbol
			LIMIT ?
		`, docset, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("list query failed: %w", err)
	}
	defer rows.Close()

	var entries []model.Entry
	for rows.Next() {
		var e model.Entry
		err := rows.Scan(&e.Docset, &e.Version, &e.Symbol, &e.Title, &e.Content, &e.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// GetEntry returns a single entry by docset and path
func (s *Searcher) GetEntry(docset, version, path string) (*model.Entry, error) {
	var e model.Entry
	err := s.db.conn.QueryRow(`
		SELECT docset, version, symbol, title, content, path
		FROM docs
		WHERE docset = ? AND version = ? AND path = ?
	`, docset, version, path).Scan(&e.Docset, &e.Version, &e.Symbol, &e.Title, &e.Content, &e.Path)

	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ListDocsets returns all installed docsets
func (s *Searcher) ListDocsets() ([]model.Docset, error) {
	rows, err := s.db.conn.Query(`
		SELECT id, slug, name, version, display_name, entry_count, mtime, installed_at
		FROM docsets
		ORDER BY name, version DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list docsets: %w", err)
	}
	defer rows.Close()

	var docsets []model.Docset
	for rows.Next() {
		var d model.Docset
		var installedAt int64
		err := rows.Scan(&d.ID, &d.Slug, &d.Name, &d.Version, &d.DisplayName, &d.EntryCount, &d.Mtime, &installedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan docset: %w", err)
		}
		docsets = append(docsets, d)
	}

	return docsets, rows.Err()
}

// escapeFTS5Query escapes special FTS5 characters and prepares the query
func escapeFTS5Query(query string) string {
	// Remove characters that have special meaning in FTS5
	query = strings.ReplaceAll(query, "\"", "")
	query = strings.ReplaceAll(query, "'", "")
	query = strings.ReplaceAll(query, "(", "")
	query = strings.ReplaceAll(query, ")", "")
	query = strings.ReplaceAll(query, "*", "")
	query = strings.ReplaceAll(query, ":", "")

	// Split into words and add prefix matching
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}

	// Add prefix matching to the last word for typeahead
	for i := range words {
		if i == len(words)-1 {
			words[i] = words[i] + "*"
		}
	}

	return strings.Join(words, " ")
}
