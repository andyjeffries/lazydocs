package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("99")  // Purple
	colorSecondary = lipgloss.Color("241") // Gray
	colorBorder    = lipgloss.Color("240") // Dark gray
	colorActive    = lipgloss.Color("99")  // Purple for active elements
	colorMuted     = lipgloss.Color("245") // Muted text

	// Tab styles
	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("255")).
			Background(colorPrimary).
			Bold(true)

	tabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder)

	// Pane styles
	paneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorActive).
			Padding(0, 1)

	// Results list styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(colorPrimary).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(colorSecondary).
			Padding(0, 1)

	statusTextStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	// Help styles
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255"))

	// Search input style
	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)
)
