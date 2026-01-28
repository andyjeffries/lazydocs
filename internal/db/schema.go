package db

const schema = `
-- Main FTS5 table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS docs USING fts5(
    docset,
    version,
    symbol,
    title,
    content,
    path,
    tokenize = 'porter unicode61'
);

-- Metadata table for installed docsets
CREATE TABLE IF NOT EXISTS docsets (
    id INTEGER PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    display_name TEXT,
    entry_count INTEGER DEFAULT 0,
    mtime INTEGER,
    installed_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Index for faster docset lookups
CREATE INDEX IF NOT EXISTS idx_docsets_name ON docsets(name);
`
