package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yeniklas/gokapi-tui/internal/api"
	"github.com/yeniklas/gokapi-tui/internal/config"
	"github.com/yeniklas/gokapi-tui/internal/tui"
	"github.com/yeniklas/gokapi-tui/internal/updater"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	updateFlag := flag.Bool("self-update", false, "update gokapi-tui to the latest release")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *updateFlag {
		if err := updater.Run(version); err != nil {
			fmt.Fprintln(os.Stderr, "update:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Create ~/.config/gokapi-tui/config.yaml with server_url and api_key.\n")
		os.Exit(1)
	}

	client := api.New(cfg.ServerURL, cfg.APIKey)
	app := tui.New(cfg, client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
