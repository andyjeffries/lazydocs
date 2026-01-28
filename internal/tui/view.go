package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content string

	switch m.mode {
	case ModeHelp:
		content = m.viewHelp()
	case ModeDocsetPicker:
		content = m.viewDocsetPicker()
	default:
		content = m.viewMain()
	}

	return content
}

func (m Model) viewMain() string {
	// Check for empty state
	if len(m.docsets) == 0 {
		return m.viewEmpty()
	}

	// Build the three sections: tabs, main content, status bar
	tabs := m.viewTabs()
	mainContent := m.viewMainContent()
	statusBar := m.viewStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, tabs, mainContent, statusBar)
}

func (m Model) viewEmpty() string {
	emptyMsg := `
  Welcome to LazyDocs!

  No documentation installed yet.

  Press 'a' to add your first docset.

  Popular choices:
    • javascript    • go           • python~3.12
    • react         • vue~3        • ruby~3.3
    • rails~8.0     • nodejs       • typescript
`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(2, 4).
		Width(60)

	centered := lipgloss.Place(
		m.width, m.height-1,
		lipgloss.Center, lipgloss.Center,
		style.Render(emptyMsg),
	)

	// Add status bar
	statusBar := m.viewStatusBar()
	return lipgloss.JoinVertical(lipgloss.Left, centered, statusBar)
}

func (m Model) viewTabs() string {
	var tabs []string

	for i, ds := range m.docsets {
		var tab string
		name := ds.DisplayName
		if ds.Version != "" {
			name = fmt.Sprintf("%s~%s", ds.Name, ds.Version)
		}
		if name == "" {
			name = ds.Slug
		}

		if i == m.activeTab {
			tab = activeTabStyle.Render(name)
		} else {
			tab = tabStyle.Render(name)
		}
		tabs = append(tabs, tab)
	}

	// Add the [+] button
	tabs = append(tabs, tabStyle.Render("[+]"))

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return tabBarStyle.Width(m.width).Render(tabRow)
}

func (m Model) viewMainContent() string {
	// Calculate dimensions
	totalWidth := m.width - 2
	resultsWidth := totalWidth / 3
	previewWidth := totalWidth - resultsWidth - 1

	// Height for main content area
	contentHeight := m.height - 5

	// Build results pane
	results := m.viewResults(resultsWidth-2, contentHeight-2)
	resultsPane := paneStyle
	if m.activePane == PaneResults {
		resultsPane = activePaneStyle
	}
	resultsBox := resultsPane.Width(resultsWidth).Height(contentHeight).Render(results)

	// Build preview pane
	preview := m.viewPreview(previewWidth-2, contentHeight-2)
	previewPane := paneStyle
	if m.activePane == PanePreview {
		previewPane = activePaneStyle
	}
	previewBox := previewPane.Width(previewWidth).Height(contentHeight).Render(preview)

	return lipgloss.JoinHorizontal(lipgloss.Top, resultsBox, previewBox)
}

func (m Model) viewResults(width, height int) string {
	var lines []string

	// Title - show search type
	title := "Results"
	if m.mode == ModeSearch && m.searchInput.Value() != "" {
		if m.globalSearch {
			title = "Global: " + m.searchInput.Value()
		} else {
			title = "Search: " + m.searchInput.Value()
		}
	}
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, "")

	// Show search input if in search mode
	if m.mode == ModeSearch {
		lines = append(lines, m.searchInput.View())
		lines = append(lines, "")
	}

	if len(m.entries) == 0 {
		if m.mode == ModeSearch {
			lines = append(lines, normalItemStyle.Render("  Type to search..."))
		} else {
			lines = append(lines, normalItemStyle.Render("  No entries"))
		}
		return strings.Join(lines, "\n")
	}

	// List entries
	visibleCount := height - 4
	if visibleCount < 1 {
		visibleCount = 1
	}

	start := 0
	if m.selectedIdx >= visibleCount {
		start = m.selectedIdx - visibleCount + 1
	}

	// Check if this is a global search (entries from multiple docsets)
	isGlobalSearch := m.mode == ModeSearch && m.searchInput.Value() != ""

	for i := start; i < len(m.entries) && i < start+visibleCount; i++ {
		entry := m.entries[i]
		line := entry.Symbol
		if line == "" {
			line = entry.Title
		}

		// Show docset name for global search results
		if isGlobalSearch && entry.Docset != "" {
			docsetTag := "[" + entry.Docset
			if entry.Version != "" {
				docsetTag += "~" + entry.Version
			}
			docsetTag += "]"
			line = line + " " + helpStyle.Render(docsetTag)
		}

		if i == m.selectedIdx {
			line = selectedItemStyle.Render("> ") + line
		} else {
			line = "  " + line
		}

		// Truncate if needed
		if lipgloss.Width(line) > width {
			line = line[:width-3] + "..."
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) viewPreview(width, height int) string {
	if len(m.entries) == 0 || m.selectedIdx >= len(m.entries) {
		return "No entry selected"
	}

	return m.preview.View()
}

func (m Model) viewStatusBar() string {
	// Left side: docset info or download progress
	var leftInfo string

	if m.downloading {
		leftInfo = fmt.Sprintf("Downloading %s... %.0f%% %s", m.downloadSlug, m.downloadPct, m.downloadStatus)
	} else if m.statusMsg != "" {
		leftInfo = m.statusMsg
	} else if len(m.docsets) > 0 && m.activeTab < len(m.docsets) {
		ds := m.docsets[m.activeTab]
		name := ds.DisplayName
		if name == "" {
			name = ds.Slug
		}
		leftInfo = fmt.Sprintf("%s • %d entries", name, ds.EntryCount)
	}

	// Debug: show mode (only if enabled in config)
	debugInfo := ""
	if m.showDebug {
		modeNames := []string{"NORMAL", "SEARCH", "HELP", "PICKER", "DELETE?"}
		modeName := "?"
		if int(m.mode) < len(modeNames) {
			modeName = modeNames[m.mode]
		}
		debugInfo = fmt.Sprintf("[%s]", modeName)
	}

	// Right side: key hints
	keyHints := []string{
		statusKeyStyle.Render("s") + statusTextStyle.Render("search"),
		statusKeyStyle.Render("/") + statusTextStyle.Render("global"),
		statusKeyStyle.Render("a") + statusTextStyle.Render("add"),
		statusKeyStyle.Render("d") + statusTextStyle.Render("delete"),
		statusKeyStyle.Render("?") + statusTextStyle.Render("help"),
		statusKeyStyle.Render("q") + statusTextStyle.Render("quit"),
	}
	rightInfo := strings.Join(keyHints, "  ")

	// Calculate spacing
	leftWidth := lipgloss.Width(leftInfo)
	rightWidth := lipgloss.Width(rightInfo)
	debugWidth := lipgloss.Width(debugInfo)
	extraSpace := 4
	if debugInfo != "" {
		extraSpace = 6
	}
	spacerWidth := m.width - leftWidth - rightWidth - debugWidth - extraSpace
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	if debugInfo != "" {
		return statusBarStyle.Width(m.width).Render(leftInfo + spacer + debugInfo + "  " + rightInfo)
	}
	return statusBarStyle.Width(m.width).Render(leftInfo + spacer + rightInfo)
}

func (m Model) viewHelp() string {
	help := `
 LazyDocs Help

 Navigation
 ──────────────────────────────────────
 j/k, ↑/↓      Move selection up/down
 h/l, ←/→      Switch panes / docsets
 Tab           Cycle panes forward
 Shift-Tab     Cycle panes backward
 Enter         Select / view in preview
 g             Scroll to top
 G             Scroll to bottom
 Ctrl+d        Scroll half-page down
 Ctrl+u        Scroll half-page up

 Actions
 ──────────────────────────────────────
 s             Search current docset
 /             Search all docsets (global)
 a             Add docset
 d             Delete selected docset
 u             Update selected docset
 ?             Toggle help
 q, Ctrl+c     Quit
 Esc           Close modal / clear search

 Press ? or Esc to close this help
`
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(50).
		Align(lipgloss.Left)

	centered := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpStyle.Render(help),
	)

	return centered
}

func (m Model) viewDocsetPicker() string {
	var content strings.Builder

	content.WriteString(" Add Documentation\n")
	content.WriteString("─────────────────────────────────────────────\n\n")

	// Show download progress if downloading
	if m.downloading {
		content.WriteString(fmt.Sprintf(" Downloading %s...\n\n", m.downloadSlug))
		content.WriteString(" Please wait, this may take a moment.\n")
		content.WriteString(" The UI will update when complete.\n")

		pickerStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Width(60)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			pickerStyle.Render(content.String()),
		)
	}

	content.WriteString(" " + m.pickerSearch.View() + "\n\n")

	filtered := m.filteredManifest()

	if len(filtered) == 0 {
		if len(m.availableDocsets) == 0 {
			content.WriteString(" Loading docsets...\n")
		} else {
			content.WriteString(" No matching docsets\n")
		}
	} else {
		// Show up to 15 entries
		maxVisible := 15
		start := 0
		if m.pickerIdx >= maxVisible {
			start = m.pickerIdx - maxVisible + 1
		}

		// Build a set of installed slugs for quick lookup
		installed := make(map[string]bool)
		for _, ds := range m.docsets {
			installed[ds.Slug] = true
		}

		for i := start; i < len(filtered) && i < start+maxVisible; i++ {
			entry := filtered[i]

			// Format: slug (display name) [size] [installed]
			line := fmt.Sprintf("%-20s %s", entry.Slug, entry.Name)
			if entry.DBSize > 0 {
				sizeMB := float64(entry.DBSize) / 1024 / 1024
				line += fmt.Sprintf(" (%.1f MB)", sizeMB)
			}

			if installed[entry.Slug] {
				line += " [installed]"
			}

			if i == m.pickerIdx {
				content.WriteString(" " + selectedItemStyle.Render("> "+line) + "\n")
			} else {
				content.WriteString("   " + line + "\n")
			}
		}
	}

	content.WriteString("\n─────────────────────────────────────────────\n")
	content.WriteString(" Enter: install  Esc: cancel")

	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(60)

	centered := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		pickerStyle.Render(content.String()),
	)

	return centered
}
