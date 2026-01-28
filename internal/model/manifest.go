package model

// ManifestEntry represents a docset in the DevDocs manifest
type ManifestEntry struct {
	Name    string `json:"name"`    // e.g., "Ruby on Rails"
	Slug    string `json:"slug"`    // e.g., "rails~8.0"
	Type    string `json:"type"`    // e.g., "rails"
	Version string `json:"version"` // e.g., "8.0"
	Release string `json:"release"` // e.g., "8.0.0"
	Mtime   int64  `json:"mtime"`   // Unix timestamp
	DBSize  int64  `json:"db_size"` // Size in bytes
}

// Manifest is the full list of available docsets
type Manifest []ManifestEntry

// ParseSlug extracts name and version from a slug like "rails~7.1"
func ParseSlug(slug string) (name, version string) {
	for i := len(slug) - 1; i >= 0; i-- {
		if slug[i] == '~' {
			return slug[:i], slug[i+1:]
		}
	}
	return slug, ""
}
