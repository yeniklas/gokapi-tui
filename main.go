package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yeniklas/gokapi-tui/internal/api"
	"github.com/yeniklas/gokapi-tui/internal/config"
	"github.com/yeniklas/gokapi-tui/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
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
