package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))
	normalStyle  = lipgloss.NewStyle()
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func renderList(files []model.GokapiFile, cursor, width int) string {
	if width < 20 {
		width = 80
	}
	nameW := width / 3
	sizeW := 10
	expiryW := 12

	header := headerStyle.Render(
		fmt.Sprintf("%-*s %-*s %-*s %s",
			nameW, "NAME",
			sizeW, "SIZE",
			expiryW, "EXPIRES",
			"DOWNLOADS",
		),
	)

	var rows []string
	rows = append(rows, header)

	for i, f := range files {
		name := truncate(f.Name, nameW)
		size := truncate(f.Size, sizeW)

		expiry := f.ExpireAtString
		if f.UnlimitedTime {
			expiry = "never"
		}
		expiry = truncate(expiry, expiryW)

		var dl string
		if f.UnlimitedDownloads {
			dl = "unlimited"
		} else {
			dl = fmt.Sprintf("%d left", f.DownloadsRemaining)
		}

		row := fmt.Sprintf("%-*s %-*s %-*s %s",
			nameW, name,
			sizeW, size,
			expiryW, expiry,
			dl,
		)

		if i == cursor {
			rows = append(rows, selectedStyle.Render("> "+row))
		} else {
			rows = append(rows, normalStyle.Render("  "+row))
		}
	}

	if len(files) == 0 {
		rows = append(rows, dimStyle.Render("  No files found. Press u to upload."))
	}

	return strings.Join(rows, "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
