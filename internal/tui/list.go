package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))
	normalStyle   = lipgloss.NewStyle()
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

// renderHelp renders shortcut hints with keys accented and descriptions dimmed.
// Input format: "key:desc  key:desc  ..." with double-space separators.
func renderHelp(s string) string {
	parts := strings.Split(s, "  ")
	out := make([]string, len(parts))
	for i, p := range parts {
		if p == "|" {
			out[i] = dimStyle.Render(p)
			continue
		}
		if idx := strings.Index(p, ":"); idx >= 0 {
			out[i] = helpKeyStyle.Render(p[:idx]) + dimStyle.Render(p[idx:])
		} else {
			out[i] = dimStyle.Render(p)
		}
	}
	return strings.Join(out, "  ")
}

const (
	sizeW     = 9
	expiryW   = 12
	dlW       = 10
	uploadedW = 11
	pwW       = 3
	// 2 (cursor) + sizeW + expiryW + dlW + uploadedW + pwW + 5 spaces between columns
	fixedW = 2 + sizeW + expiryW + dlW + uploadedW + pwW + 5
)

func renderList(files []model.GokapiFile, cursor, width int) string {
	if width < 20 {
		width = 80
	}
	nameW := width - fixedW
	if nameW < 8 {
		nameW = 8
	}

	header := headerStyle.Render(
		fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %s",
			nameW, "NAME",
			sizeW, "SIZE",
			expiryW, "EXPIRES",
			dlW, "DOWNLOADS",
			uploadedW, "UPLOADED",
			"PW",
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

		uploaded := ""
		if f.UploadDate > 0 {
			uploaded = time.Unix(f.UploadDate, 0).Format("2006-01-02")
		}

		pw := ""
		if f.IsPasswordProtected {
			pw = "yes"
		}

		row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
			nameW, name,
			sizeW, size,
			expiryW, expiry,
			dlW, dl,
			uploadedW, uploaded,
			pw,
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
