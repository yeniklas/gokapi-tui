package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	tabUploads      = 0
	tabFileRequests = 1
	tabCount        = 2
)

var tabNames = []string{"Uploads", "File Requests"}

var (
	activeTabStyle   = lipgloss.NewStyle().Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func renderTabBar(active, _ int) string {
	parts := make([]string, tabCount)
	for i, name := range tabNames {
		if i == active {
			parts[i] = activeTabStyle.Render("[" + name + "]")
		} else {
			parts[i] = inactiveTabStyle.Render(name)
		}
	}
	return strings.Join(parts, "  ")
}
