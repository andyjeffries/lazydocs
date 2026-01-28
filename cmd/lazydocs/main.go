package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazydocs/lazydocs/internal/app"
	"github.com/lazydocs/lazydocs/internal/tui"
	"github.com/muesli/termenv"
)

func init() {
	// Force TrueColor profile to avoid terminal color queries
	// which can interfere with key input on some terminals (e.g., Ghostty)
	lipgloss.SetColorProfile(termenv.TrueColor)
}

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		runTUI("")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: lazydocs install <docset>")
			os.Exit(1)
		}
		installDocset(os.Args[2])

	case "remove", "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: lazydocs remove <docset>")
			os.Exit(1)
		}
		removeDocset(os.Args[2])

	case "list":
		listDocsets()

	case "search", "available":
		filter := ""
		if len(os.Args) >= 3 {
			filter = os.Args[2]
		}
		searchAvailable(filter)

	case "update":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: lazydocs update <docset|all>")
			os.Exit(1)
		}
		updateDocset(os.Args[2])

	case "--lookup", "-l":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: lazydocs --lookup <symbol>")
			os.Exit(1)
		}
		runTUI(os.Args[2])

	case "--version", "-v":
		fmt.Printf("lazydocs %s\n", version)

	case "--help", "-h", "help":
		printHelp()

	default:
		// Check if it looks like a lookup
		if strings.HasPrefix(cmd, "-") {
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", cmd)
			printHelp()
			os.Exit(1)
		}
		// Treat as lookup query
		runTUI(cmd)
	}
}

func runTUI(lookup string) {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	p := tea.NewProgram(
		tui.NewWithApp(application, lookup),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func installDocset(slug string) {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	fmt.Printf("Installing %s...\n", slug)

	err = application.InstallDocset(slug, func(downloaded, total int64, status string) {
		if total > 0 {
			pct := float64(downloaded) / float64(total) * 100
			fmt.Printf("\r%s %.1f%% (%d/%d bytes)", status, pct, downloaded, total)
		} else {
			fmt.Printf("\r%s %d bytes", status, downloaded)
		}
	})

	fmt.Println()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully installed %s\n", slug)
}

func removeDocset(slug string) {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	if err := application.RemoveDocset(slug); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed %s\n", slug)
}

func listDocsets() {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	docsets, err := application.ListInstalledDocsets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(docsets) == 0 {
		fmt.Println("No docsets installed. Use 'lazydocs install <docset>' to install one.")
		fmt.Println("\nPopular docsets: javascript, go, python~3.12, ruby~3.3, react, vue~3")
		return
	}

	fmt.Println("Installed docsets:")
	for _, ds := range docsets {
		fmt.Printf("  %-20s %s (%d entries)\n", ds.Slug, ds.DisplayName, ds.EntryCount)
	}
}

func updateDocset(slug string) {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	if slug == "all" {
		docsets, err := application.ListInstalledDocsets()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, ds := range docsets {
			fmt.Printf("Updating %s...\n", ds.Slug)
			if err := application.InstallDocset(ds.Slug, nil); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", ds.Slug, err)
			}
		}
	} else {
		fmt.Printf("Updating %s...\n", slug)
		if err := application.InstallDocset(slug, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Update complete")
}

func searchAvailable(filter string) {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	fmt.Println("Fetching available docsets...")

	manifest, err := application.ListAvailableDocsets(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	filter = strings.ToLower(filter)
	count := 0

	fmt.Printf("\nAvailable docsets")
	if filter != "" {
		fmt.Printf(" (filter: %q)", filter)
	}
	fmt.Println(":\n")

	for _, entry := range manifest {
		// Filter if specified
		if filter != "" {
			if !strings.Contains(strings.ToLower(entry.Slug), filter) &&
				!strings.Contains(strings.ToLower(entry.Name), filter) {
				continue
			}
		}

		sizeMB := float64(entry.DBSize) / 1024 / 1024
		fmt.Printf("  %-25s %-30s (%.1f MB)\n", entry.Slug, entry.Name, sizeMB)
		count++
	}

	fmt.Printf("\n%d docsets found\n", count)
	fmt.Println("\nInstall with: lazydocs install <slug>")
}

func printHelp() {
	fmt.Print(`LazyDocs - Lazygit-style TUI for browsing DevDocs documentation

Usage:
  lazydocs                    Open the TUI
  lazydocs <symbol>           Open TUI and search for symbol
  lazydocs --lookup <symbol>  Same as above

Commands:
  install <docset>     Install a docset (e.g., javascript, go, python~3.12)
  remove <docset>      Remove an installed docset
  update <docset|all>  Update a docset or all docsets
  list                 List installed docsets
  search [filter]      Search available docsets to install

Options:
  --help, -h           Show this help
  --version, -v        Show version

Examples:
  lazydocs install javascript
  lazydocs install python~3.12
  lazydocs search python        Search available Python docsets
  lazydocs search               List all available docsets
  lazydocs list

Popular docsets: javascript, go, python~3.12, ruby~3.3, react, vue~3, rails~8.0
`)
}
