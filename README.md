# LazyDocs

A Lazygit-style TUI for browsing offline DevDocs documentation.

## Features

- **Offline documentation** - Download and browse docs without internet
- **Fast full-text search** - SQLite FTS5 with BM25 ranking
- **Markdown rendering** - Beautiful terminal rendering with syntax highlighting
- **Multiple docsets** - Install and switch between documentation sets
- **Vim-style navigation** - `j/k`, `h/l`, `g/G`, `Ctrl+d/u`
- **Neovim integration** - Use as a floating window inside Neovim

## Installation

### From Source

Requires Go 1.21+ with CGO enabled.

```bash
git clone https://github.com/lazydocs/lazydocs
cd lazydocs
make install
```

### Binary Releases

Download from the [releases page](https://github.com/lazydocs/lazydocs/releases).

## Usage

### Standalone

```bash
# Open the TUI
lazydocs

# Search for a symbol
lazydocs Array.map

# Install a docset
lazydocs install javascript

# List installed docsets
lazydocs list

# Update a docset
lazydocs update javascript

# Remove a docset
lazydocs remove javascript
```

### Neovim

See [lazydocs.nvim](https://github.com/lazydocs/lazydocs.nvim) for Neovim integration.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Move selection up/down |
| `h/l` or `←/→` | Switch docset tabs |
| `Tab` | Cycle panes forward |
| `Shift-Tab` | Cycle panes backward |
| `Enter` | View in preview pane |
| `g` | Scroll to top |
| `G` | Scroll to bottom |
| `Ctrl+d` | Scroll half-page down |
| `Ctrl+u` | Scroll half-page up |

### Actions

| Key | Action |
|-----|--------|
| `/` | Search current docset |
| `a` | Add docset |
| `d` | Delete selected docset |
| `u` | Update selected docset |
| `?` | Toggle help |
| `q` | Quit |
| `Esc` | Close modal / cancel |

## Popular Docsets

```bash
lazydocs install javascript
lazydocs install go
lazydocs install python~3.12
lazydocs install ruby~3.3
lazydocs install react
lazydocs install vue~3
lazydocs install rails~8.0
lazydocs install nodejs
lazydocs install typescript
```

## Configuration

LazyDocs stores data in `~/.local/share/lazydocs/`:

```
lazydocs/
├── docs/           # Downloaded docsets
├── index.sqlite    # Search index
└── manifest.json   # Cached docset list
```

## Building

```bash
# Build
make build

# Run tests
make test

# Install to ~/.local/bin
make install

# Clean
make clean
```

## Credits

- Documentation from [DevDocs](https://devdocs.io)
- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), and [Glamour](https://github.com/charmbracelet/glamour)
- Inspired by [Lazygit](https://github.com/jesseduffield/lazygit)

## License

MIT
