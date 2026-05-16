package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/yeniklas/gokapi-tui/internal/model"
)

const (
	frNameW      = 0 // dynamic
	frExpiryW    = 12
	frMaxFilesW  = 10
	frUploadedW  = 10
	frTotalSizeW = 12
	// 2 (cursor) + frExpiryW + frMaxFilesW + frUploadedW + frTotalSizeW + 4 spaces
	frFixedW = 2 + frExpiryW + frMaxFilesW + frUploadedW + frTotalSizeW + 4
)

func renderFRList(frs []model.FileRequest, cursor, width int) string {
	if width < 20 {
		width = 80
	}
	nameW := width - frFixedW
	if nameW < 8 {
		nameW = 8
	}

	header := headerStyle.Render(
		fmt.Sprintf("  %-*s %-*s %-*s %-*s %s",
			nameW, "NAME",
			frExpiryW, "EXPIRES",
			frMaxFilesW, "MAX FILES",
			frUploadedW, "UPLOADED",
			"TOTAL SIZE",
		),
	)

	var rows []string
	rows = append(rows, header)

	for i, fr := range frs {
		name := truncate(fr.Name, nameW)

		expiry := "never"
		if fr.Expiry > 0 {
			expiry = time.Unix(fr.Expiry, 0).Format("2006-01-02")
		}
		expiry = truncate(expiry, frExpiryW)

		maxFiles := "unlimited"
		if fr.MaxFiles > 0 {
			maxFiles = fmt.Sprintf("%d", fr.MaxFiles)
		}

		uploaded := fmt.Sprintf("%d", fr.UploadedFiles)

		totalSize := formatBytes(fr.TotalFileSize)

		row := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
			nameW, name,
			frExpiryW, expiry,
			frMaxFilesW, maxFiles,
			frUploadedW, uploaded,
			totalSize,
		)

		if i == cursor {
			rows = append(rows, selectedStyle.Render("> "+row))
		} else {
			rows = append(rows, normalStyle.Render("  "+row))
		}
	}

	if len(frs) == 0 {
		rows = append(rows, dimStyle.Render("  No file requests found. Press n to create one."))
	}

	return strings.Join(rows, "\n")
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
