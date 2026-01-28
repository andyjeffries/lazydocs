package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/lazydocs/lazydocs/internal/app"
	"github.com/lazydocs/lazydocs/internal/model"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeHelp
	ModeDocsetPicker
	ModeDeleteConfirm
)

// Pane represents which pane is focused
type Pane int

const (
	PaneResults Pane = iota
	PanePreview
)

// Model is the main Bubbletea model
type Model struct {
	// App reference
	app *app.App

	// Dimensions
	width  int
	height int

	// State
	mode        Mode
	activePane  Pane
	activeTab   int

	// Data
	docsets     []model.Docset
	entries     []model.Entry
	selectedIdx int

	// Components
	searchInput textinput.Model
	preview     viewport.Model

	// Navigation history
	history []string

	// Status
	statusMsg   string
	initialQuery string

	// Available docsets (for picker)
	availableDocsets model.Manifest
	pickerIdx        int
	pickerSearch     textinput.Model

	// Download progress
	downloading    bool
	downloadSlug   string
	downloadPct    float64
	downloadStatus string

	// Config
	theme     string // "dark", "light", "dracula", "notty"
	showDebug bool

	// Debug
	lastKey string


	// Track if search filter is active (for Escape to clear)
	searchActive    bool
	globalSearch    bool // true = search all docsets, false = current docset only
	lastRenderedIdx int  // Track which entry was last rendered in preview
}

// New creates a new Model with demo data (for testing without app)
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.Prompt = "/ "
	ti.PromptStyle = searchPromptStyle
	ti.CharLimit = 256
	ti.Width = 40

	pickerInput := textinput.New()
	pickerInput.Placeholder = "Type to filter..."
	pickerInput.Prompt = "> "
	pickerInput.CharLimit = 64
	pickerInput.Width = 50

	vp := viewport.New(0, 0)

	// Demo docsets for testing
	docsets := []model.Docset{
		{Slug: "javascript", Name: "javascript", DisplayName: "JavaScript", EntryCount: 2847},
		{Slug: "go", Name: "go", DisplayName: "Go", EntryCount: 1523},
	}

	// Demo entries
	entries := []model.Entry{
		{Symbol: "Array.map", Title: "Array.prototype.map()", Content: "# Array.prototype.map()\n\nThe `map()` method creates a new array populated with the results of calling a provided function on every element in the calling array."},
		{Symbol: "Array.filter", Title: "Array.prototype.filter()", Content: "# Array.prototype.filter()\n\nThe `filter()` method creates a shallow copy of a portion of a given array."},
	}

	return Model{
		mode:         ModeNormal,
		activePane:   PaneResults,
		activeTab:    0,
		docsets:      docsets,
		entries:      entries,
		selectedIdx:  0,
		searchInput:  ti,
		pickerSearch: pickerInput,
		preview:      vp,
		statusMsg:    "",
	}
}

// NewWithApp creates a new Model connected to the app
func NewWithApp(application *app.App, lookup string) Model {
	m := New()
	m.app = application
	m.initialQuery = lookup

	// Load config settings
	cfg := application.Config()
	m.theme = cfg.Theme
	if m.theme == "" {
		m.theme = "dark"
	}
	m.showDebug = cfg.UI.ShowDebug

	// Load installed docsets
	docsets, err := application.ListInstalledDocsets()
	if err == nil && len(docsets) > 0 {
		m.docsets = docsets

		// Load entries for first docset
		entries, err := application.ListEntries(docsets[0].Name, docsets[0].Version, 100)
		if err == nil {
			m.entries = entries
		}
	} else {
		m.docsets = nil
		m.entries = nil
	}

	// If lookup query provided, start in search mode
	if lookup != "" {
		m.searchInput.SetValue(lookup)
		m.mode = ModeSearch
		m.searchInput.Focus()
	}

	return m
}

// currentDocset returns the currently active docset
func (m Model) currentDocset() *model.Docset {
	if m.activeTab >= 0 && m.activeTab < len(m.docsets) {
		return &m.docsets[m.activeTab]
	}
	return nil
}
