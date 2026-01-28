package model

// Entry represents a documentation entry
type Entry struct {
	Docset  string // e.g., "rails"
	Version string // e.g., "7.1" (empty for unversioned)
	Symbol  string // e.g., "ActiveRecord::Base"
	Title   string // Display title
	Content string // Full markdown content
	Path    string // Original path in docset
}

// SearchResult represents a search result with ranking info
type SearchResult struct {
	Entry
	Rank    float64 // BM25 rank score
	Snippet string  // Highlighted snippet from content
}
