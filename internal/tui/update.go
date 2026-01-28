package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/lazydocs/lazydocs/internal/data"
	"github.com/lazydocs/lazydocs/internal/model"
)

// Messages for async operations
type startDownloadMsg struct {
	slug string
}

type docsetInstalledMsg struct {
	slug string
	err  error
}

type downloadProgressMsg struct {
	downloaded int64
	total      int64
	status     string
}

type searchResultsMsg struct {
	results []model.SearchResult
	err     error
}

type entriesLoadedMsg struct {
	entries []model.Entry
	err     error
}

type manifestLoadedMsg struct {
	manifest model.Manifest
	err      error
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// If we have a search query, execute it
	if m.initialQuery != "" && m.app != nil {
		cmds = append(cmds, m.doSearch(m.initialQuery))
	}

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateDimensions()

	case startDownloadMsg:
		// Now actually start the download (UI has been updated to show downloading)
		return m, m.installDocset(msg.slug)

	case downloadProgressMsg:
		m.downloadPct = float64(msg.downloaded) / float64(msg.total) * 100
		m.downloadStatus = msg.status

	case docsetInstalledMsg:
		m.downloading = false
		m.downloadSlug = ""
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = "Installed successfully"
			// Reload docsets
			if m.app != nil {
				docsets, err := m.app.ListInstalledDocsets()
				if err == nil {
					m.docsets = docsets
					// Switch to the new docset
					for i, ds := range docsets {
						if ds.Slug == msg.slug {
							m.activeTab = i
							// Load entries
							cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
							break
						}
					}
				}
			}
		}
		m.mode = ModeNormal

	case searchResultsMsg:
		if msg.err == nil {
			m.entries = make([]model.Entry, len(msg.results))
			for i, r := range msg.results {
				m.entries[i] = r.Entry
			}
			m.selectedIdx = 0
			m.lastRenderedIdx = -1 // Force preview update
			m = m.updatePreviewContent()
		}

	case entriesLoadedMsg:
		if msg.err == nil {
			m.entries = msg.entries
			m.selectedIdx = 0
			m.lastRenderedIdx = -1 // Force preview update
			m = m.updatePreviewContent()
		}

	case manifestLoadedMsg:
		if msg.err == nil {
			m.availableDocsets = msg.manifest
		}

	case tea.KeyMsg:
		// Track last key for debugging
		m.lastKey = msg.String()

		// Filter out terminal response sequences (they often contain semicolons or "rgb")
		keyStr := msg.String()
		if strings.Contains(keyStr, ";") || strings.Contains(keyStr, "rgb") || len(keyStr) > 10 {
			// This looks like a terminal escape sequence response, ignore it
			return m, nil
		}

		// Always allow Ctrl+C to quit
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle mode-specific keys first
		switch m.mode {
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeHelp:
			return m.updateHelp(msg)
		case ModeDocsetPicker:
			return m.updateDocsetPicker(msg)
		case ModeDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		default:
			return m.updateNormal(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateDimensions() Model {
	// Calculate preview pane dimensions
	previewWidth := (m.width * 2 / 3) - 4
	previewHeight := m.height - 6

	if previewWidth > 0 && previewHeight > 0 {
		m.preview.Width = previewWidth
		m.preview.Height = previewHeight
	}

	// Update preview content
	m = m.updatePreviewContent()

	return m
}

func (m Model) updatePreviewContent() Model {
	if len(m.entries) == 0 || m.selectedIdx >= len(m.entries) {
		m.preview.SetContent("No entry selected")
		m.lastRenderedIdx = -1
		return m
	}

	// Only re-render if selection changed
	if m.selectedIdx == m.lastRenderedIdx {
		return m
	}

	entry := m.entries[m.selectedIdx]

	// Render markdown content with configured theme
	theme := m.theme
	if theme == "" {
		theme = "dark"
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(theme),
		glamour.WithWordWrap(m.preview.Width),
	)

	var content string
	if err != nil {
		content = entry.Content
	} else {
		rendered, err := renderer.Render(entry.Content)
		if err != nil {
			content = entry.Content
		} else {
			content = rendered
		}
	}

	m.preview.SetContent(content)
	m.preview.GotoTop()
	m.lastRenderedIdx = m.selectedIdx

	return m
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Escape):
		// Clear search filter if active
		if m.searchActive {
			m.searchActive = false
			m.searchInput.SetValue("")
			if ds := m.currentDocset(); ds != nil && m.app != nil {
				cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
			}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case key.Matches(msg, keys.Help):
		m.mode = ModeHelp
		return m, nil

	case key.Matches(msg, keys.Search):
		m.mode = ModeSearch
		m.globalSearch = true
		m.searchInput.SetValue("")
		m.searchInput.Placeholder = "Search all docsets..."
		m.searchInput.Focus()
		return m, nil

	case key.Matches(msg, keys.LocalSearch):
		if m.currentDocset() == nil {
			return m, nil // No docset to search
		}
		m.mode = ModeSearch
		m.globalSearch = false
		m.searchInput.SetValue("")
		m.searchInput.Placeholder = "Search current docset..."
		m.searchInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Add):
		m.mode = ModeDocsetPicker
		m.pickerIdx = 0
		m.pickerSearch.SetValue("")
		m.pickerSearch.Focus() // Don't use the returned command
		// Load manifest if not loaded
		if m.app != nil && len(m.availableDocsets) == 0 {
			cmds = append(cmds, m.loadManifest())
		}
		return m, tea.Batch(cmds...)

	case key.Matches(msg, keys.Delete):
		// Show delete confirmation
		if ds := m.currentDocset(); ds != nil {
			m.mode = ModeDeleteConfirm
			m.statusMsg = "Delete " + ds.Slug + "? (y/n)"
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.activePane == PaneResults && m.selectedIdx > 0 {
			m.selectedIdx--
			m = m.updatePreviewContent()
		} else if m.activePane == PanePreview {
			m.preview.LineUp(3)
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.activePane == PaneResults && m.selectedIdx < len(m.entries)-1 {
			m.selectedIdx++
			m = m.updatePreviewContent()
		} else if m.activePane == PanePreview {
			m.preview.LineDown(3)
		}
		return m, nil

	case key.Matches(msg, keys.Left):
		if m.activePane == PaneResults && m.activeTab > 0 {
			m.activeTab--
			if ds := m.currentDocset(); ds != nil && m.app != nil {
				cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
			}
		} else {
			m.activePane = PaneResults
		}
		return m, tea.Batch(cmds...)

	case key.Matches(msg, keys.Right):
		if m.activePane == PaneResults && m.activeTab < len(m.docsets)-1 {
			m.activeTab++
			if ds := m.currentDocset(); ds != nil && m.app != nil {
				cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
			}
		} else {
			m.activePane = PanePreview
		}
		return m, tea.Batch(cmds...)

	case key.Matches(msg, keys.Tab):
		if m.activePane == PaneResults {
			m.activePane = PanePreview
		} else {
			m.activePane = PaneResults
		}
		return m, nil

	case key.Matches(msg, keys.ShiftTab):
		if m.activePane == PanePreview {
			m.activePane = PaneResults
		} else {
			m.activePane = PanePreview
		}
		return m, nil

	case key.Matches(msg, keys.Top):
		if m.activePane == PaneResults {
			m.selectedIdx = 0
		} else {
			m.preview.GotoTop()
		}
		return m, nil

	case key.Matches(msg, keys.Bottom):
		if m.activePane == PaneResults {
			if len(m.entries) > 0 {
				m.selectedIdx = len(m.entries) - 1
			}
		} else {
			m.preview.GotoBottom()
		}
		return m, nil

	case key.Matches(msg, keys.HalfDown):
		if m.activePane == PanePreview {
			m.preview.HalfViewDown()
		}
		return m, nil

	case key.Matches(msg, keys.HalfUp):
		if m.activePane == PanePreview {
			m.preview.HalfViewUp()
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.activePane = PanePreview
		return m, nil
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = ModeNormal
		m.searchInput.Blur()
		m.searchActive = false
		query := m.searchInput.Value()
		m.searchInput.SetValue("")
		// Reload entries without filter
		if query != "" {
			if ds := m.currentDocset(); ds != nil && m.app != nil {
				cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
			} else {
				m.entries = nil
			}
		}
		return m, tea.Batch(cmds...)

	case key.Matches(msg, keys.Enter):
		m.mode = ModeNormal
		m.searchInput.Blur()
		// Keep search results visible, mark search as active so Escape can clear it
		if m.searchInput.Value() != "" {
			m.searchActive = true
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		// Navigate results while searching
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		// Navigate results while searching
		if m.selectedIdx < len(m.entries)-1 {
			m.selectedIdx++
		}
		return m, nil
	}

	// Pass other keys to the text input
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	// Real-time search
	query := m.searchInput.Value()
	if query != "" && m.app != nil {
		if m.globalSearch {
			cmds = append(cmds, m.doGlobalSearch(query))
		} else {
			cmds = append(cmds, m.doLocalSearch(query))
		}
	} else if m.app != nil {
		// Empty query - reload current docset entries
		if ds := m.currentDocset(); ds != nil {
			cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Help), key.Matches(msg, keys.Quit):
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) updateDocsetPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = ModeNormal
		m.pickerSearch.Blur()
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.pickerIdx > 0 {
			m.pickerIdx--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		filtered := m.filteredManifest()
		if m.pickerIdx < len(filtered)-1 {
			m.pickerIdx++
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		filtered := m.filteredManifest()
		if m.pickerIdx < len(filtered) && m.app != nil {
			entry := filtered[m.pickerIdx]
			m.downloading = true
			m.downloadSlug = entry.Slug
			m.downloadPct = 0
			m.downloadStatus = "Starting..."
			// Return first to update UI, then start download on next tick
			cmds = append(cmds, func() tea.Msg {
				return startDownloadMsg{slug: entry.Slug}
			})
		}
		return m, tea.Batch(cmds...)
	}

	// Pass other keys to the text input
	prevValue := m.pickerSearch.Value()
	var cmd tea.Cmd
	m.pickerSearch, cmd = m.pickerSearch.Update(msg)
	cmds = append(cmds, cmd)

	// Reset selection when filter changes
	if m.pickerSearch.Value() != prevValue {
		m.pickerIdx = 0
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "y", "Y":
		// Confirmed delete
		if ds := m.currentDocset(); ds != nil && m.app != nil {
			if err := m.app.RemoveDocset(ds.Slug); err == nil {
				m.statusMsg = "Deleted " + ds.Slug
				// Reload docsets
				docsets, _ := m.app.ListInstalledDocsets()
				m.docsets = docsets
				if m.activeTab >= len(m.docsets) {
					m.activeTab = len(m.docsets) - 1
				}
				if m.activeTab < 0 {
					m.activeTab = 0
				}
				// Load entries for new active docset
				if ds := m.currentDocset(); ds != nil {
					cmds = append(cmds, m.loadEntries(ds.Name, ds.Version))
				} else {
					m.entries = nil
				}
			} else {
				m.statusMsg = "Error: " + err.Error()
			}
		}
		m.mode = ModeNormal
		return m, tea.Batch(cmds...)

	case "n", "N", "esc":
		m.mode = ModeNormal
		m.statusMsg = ""
		return m, nil

	default:
		// Ignore other keys
		return m, nil
	}
}

// filteredManifest returns manifest entries matching the picker search
func (m Model) filteredManifest() []model.ManifestEntry {
	query := strings.ToLower(m.pickerSearch.Value())
	if query == "" {
		return m.availableDocsets
	}

	var filtered []model.ManifestEntry
	for _, entry := range m.availableDocsets {
		if strings.Contains(strings.ToLower(entry.Slug), query) ||
			strings.Contains(strings.ToLower(entry.Name), query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Commands

func (m Model) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return searchResultsMsg{err: nil}
		}

		ds := m.currentDocset()
		docset := ""
		version := ""
		if ds != nil {
			docset = ds.Name
			version = ds.Version
		}

		results, err := m.app.Search(query, docset, version, 100)
		return searchResultsMsg{results: results, err: err}
	}
}

func (m Model) doGlobalSearch(query string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return searchResultsMsg{err: nil}
		}

		// Search across ALL docsets (empty docset/version = global)
		results, err := m.app.Search(query, "", "", 100)
		return searchResultsMsg{results: results, err: err}
	}
}

func (m Model) doLocalSearch(query string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return searchResultsMsg{err: nil}
		}

		ds := m.currentDocset()
		if ds == nil {
			return searchResultsMsg{err: nil}
		}

		// Search only current docset
		results, err := m.app.Search(query, ds.Name, ds.Version, 100)
		return searchResultsMsg{results: results, err: err}
	}
}

func (m Model) loadEntries(docset, version string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return entriesLoadedMsg{err: nil}
		}

		entries, err := m.app.ListEntries(docset, version, 100)
		return entriesLoadedMsg{entries: entries, err: err}
	}
}

func (m Model) loadManifest() tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return manifestLoadedMsg{err: nil}
		}

		manifest, err := m.app.ListAvailableDocsets(false)
		return manifestLoadedMsg{manifest: manifest, err: err}
	}
}

func (m Model) installDocset(slug string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return docsetInstalledMsg{slug: slug, err: nil}
		}

		err := m.app.InstallDocset(slug, func(downloaded, total int64, status string) {
			// Note: We can't send tea.Msg from here in a simple way
			// The progress will be shown via the downloading state
			_ = data.ProgressCallback(func(d, t int64, s string) {})
		})

		return docsetInstalledMsg{slug: slug, err: err}
	}
}
