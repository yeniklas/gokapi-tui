package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

var (
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1)
	detailLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

func renderDetail(f model.GokapiFile, width, height int) string {
	if width < 10 {
		width = 30
	}

	field := func(label, value string) string {
		return detailLabelStyle.Render(label+": ") + value
	}

	uploaded := ""
	if f.UploadDate > 0 {
		uploaded = time.Unix(f.UploadDate, 0).Format("2006-01-02 15:04")
	}

	expires := f.ExpireAtString
	if f.UnlimitedTime {
		expires = "never"
	}

	dl := fmt.Sprintf("%d remaining / %d total", f.DownloadsRemaining, f.DownloadCount)
	if f.UnlimitedDownloads {
		dl = fmt.Sprintf("unlimited (%d downloaded)", f.DownloadCount)
	}

	pwProtected := "no"
	if f.IsPasswordProtected {
		pwProtected = "yes"
	}

	lines := []string{
		headerStyle.Render("File Details"),
		"",
		field("Name", f.Name),
		field("Size", f.Size),
		field("Type", f.ContentType),
		field("Uploaded", uploaded),
		field("Expires", expires),
		field("Downloads", dl),
		field("Password", pwProtected),
		"",
		detailLabelStyle.Render("Share URL:"),
		wrapURL(f.UrlDownload, width-4),
	}

	content := strings.Join(lines, "\n")
	return detailBorderStyle.Width(width - 2).Render(content)
}

func wrapURL(url string, width int) string {
	if width <= 0 || len(url) <= width {
		return url
	}
	var parts []string
	for len(url) > width {
		parts = append(parts, url[:width])
		url = url[width:]
	}
	if url != "" {
		parts = append(parts, url)
	}
	return strings.Join(parts, "\n")
}
