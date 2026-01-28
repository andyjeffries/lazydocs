package model

import "time"

// Docset represents an installed documentation set
type Docset struct {
	ID          int64
	Slug        string // e.g., "rails~7.1"
	Name        string // e.g., "rails"
	Version     string // e.g., "7.1" (empty for unversioned)
	DisplayName string // e.g., "Ruby on Rails"
	EntryCount  int
	Mtime       int64     // DevDocs modification time
	InstalledAt time.Time // When we installed it
}

// FullSlug returns the slug with version if applicable
func (d Docset) FullSlug() string {
	if d.Version != "" {
		return d.Name + "~" + d.Version
	}
	return d.Name
}
