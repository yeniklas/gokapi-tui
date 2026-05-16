package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

const metricsHeight = 6 // lines consumed by the metrics panel

func renderStatus(status model.SystemStatus, logLines []string, scroll, width, height int) string {
	var b strings.Builder

	label := lipgloss.NewStyle().Width(16)
	value := lipgloss.NewStyle().Bold(true)
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Repeat("─", width))

	// ── Metrics panel ────────────────────────────────────────────────────────
	b.WriteString(headerStyle.Render("SYSTEM STATUS") + "\n\n")

	mem := fmt.Sprintf("%s / %s (%d%%)", formatBytes(status.MemoryUsed), formatBytes(status.MemoryTotal), status.MemoryUsagePercentage)
	disk := fmt.Sprintf("%s / %s (%d%%)", formatBytes(status.DiskUsed), formatBytes(status.DiskTotal), status.DiskUsagePercentage)

	rows := [][2]string{
		{"Uptime:", fmtUptime(status.Uptime)},
		{"CPU:", fmt.Sprintf("%d%%", status.CpuLoad)},
		{"Memory:", mem},
		{"Disk:", disk},
		{"Active files:", fmt.Sprintf("%d", status.ActiveFiles)},
		{"Traffic served:", formatBytes(status.DataServed)},
	}
	for _, row := range rows {
		b.WriteString(label.Render(row[0]) + value.Render(row[1]) + "\n")
	}

	b.WriteString("\n" + sep + "\n")

	// ── Logs panel ───────────────────────────────────────────────────────────
	b.WriteString(headerStyle.Render("LOGS") + "\n\n")

	if len(logLines) == 0 {
		b.WriteString(dimStyle.Render("  No logs available."))
		return b.String()
	}

	// calculate how many lines fit below the metrics + separators
	reserved := metricsHeight + len(rows) + 4 // header, blank, sep, logs header, blank
	visible := height - reserved
	if visible < 1 {
		visible = 10
	}

	end := scroll + visible
	if end > len(logLines) {
		end = len(logLines)
	}
	for _, line := range logLines[scroll:end] {
		b.WriteString("  " + line + "\n")
	}

	if len(logLines) > visible {
		shown := fmt.Sprintf("%d-%d / %d", scroll+1, end, len(logLines))
		b.WriteString("\n" + dimStyle.Render("  "+shown))
	}

	return b.String()
}

func fmtUptime(secs int64) string {
	days := secs / 86400
	secs %= 86400
	hours := secs / 3600
	secs %= 3600
	mins := secs / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
