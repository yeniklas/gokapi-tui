package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

const (
	frFieldName     = iota
	frFieldNotes
	frFieldExpiry
	frFieldMaxFiles
	frFieldMaxSize
	frFieldCount
)

func newFRInputs() []textinput.Model {
	inputs := make([]textinput.Model, frFieldCount)

	inputs[frFieldName] = textinput.New()
	inputs[frFieldName].Placeholder = "Request name"
	inputs[frFieldName].Focus()

	inputs[frFieldNotes] = textinput.New()
	inputs[frFieldNotes].Placeholder = "Optional notes"

	inputs[frFieldExpiry] = textinput.New()
	inputs[frFieldExpiry].Placeholder = "Days until expiry (0 = never)"

	inputs[frFieldMaxFiles] = textinput.New()
	inputs[frFieldMaxFiles].Placeholder = "Max files (0 = unlimited)"

	inputs[frFieldMaxSize] = textinput.New()
	inputs[frFieldMaxSize].Placeholder = "Max size MB (0 = unlimited)"

	return inputs
}

func renderFRForm(inputs []textinput.Model, activeField int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	labelStyle := lipgloss.NewStyle().Width(24)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString(titleStyle.Render("Create File Request") + "\n\n")

	labels := []string{"Name:", "Notes:", "Expiry (days):", "Max files:", "Max size (MB):"}
	for i, inp := range inputs {
		label := labelStyle.Render(labels[i])
		if i == activeField {
			b.WriteString(label + inp.View() + "\n")
		} else {
			b.WriteString(label + dim.Render(inp.Value()) + "\n")
		}
	}

	b.WriteString("\n" + renderHelp("tab: next field  |  enter: create  |  esc: cancel"))
	return b.String()
}

// parseFRParams reads the form inputs and builds CreateFileRequestParams.
func parseFRParams(inputs []textinput.Model) model.CreateFileRequestParams {
	p := model.CreateFileRequestParams{}
	p.Name = strings.TrimSpace(inputs[frFieldName].Value())
	p.Notes = strings.TrimSpace(inputs[frFieldNotes].Value())

	if days := parseInt(inputs[frFieldExpiry].Value()); days > 0 {
		p.ExpiryAt = time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()
	}
	p.MaxFiles = parseInt(inputs[frFieldMaxFiles].Value())
	p.MaxSize = parseInt(inputs[frFieldMaxSize].Value())
	return p
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		v = v*10 + int(c-'0')
	}
	return v
}
