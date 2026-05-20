package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	fieldDownloads = iota
	fieldExpiry
	fieldPassword
	fieldCount
)

func newUploadInputs(allowedDownloads, expiryDays int, password string) []textinput.Model {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldDownloads] = textinput.New()
	inputs[fieldDownloads].Placeholder = "1"
	inputs[fieldDownloads].SetValue(fmt.Sprintf("%d", allowedDownloads))
	inputs[fieldDownloads].Focus()

	inputs[fieldExpiry] = textinput.New()
	inputs[fieldExpiry].Placeholder = "7"
	inputs[fieldExpiry].SetValue(fmt.Sprintf("%d", expiryDays))

	inputs[fieldPassword] = textinput.New()
	inputs[fieldPassword].Placeholder = "leave empty for none"
	inputs[fieldPassword].SetValue(password)
	inputs[fieldPassword].EchoMode = textinput.EchoPassword

	return inputs
}

func renderUploadForm(inputs []textinput.Model, activeField int, selectedPath string) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	labelStyle := lipgloss.NewStyle().Width(22)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString(titleStyle.Render("Upload File") + "\n\n")
	b.WriteString(labelStyle.Render("File: ") + selectedPath + "\n\n")

	labels := []string{"Allowed Downloads:", "Expiry (days):", "Password:"}
	for i, inp := range inputs {
		label := labelStyle.Render(labels[i])
		if i == activeField {
			b.WriteString(label + inp.View() + "\n")
		} else {
			b.WriteString(label + dimStyle.Render(inp.Value()) + "\n")
		}
	}

	b.WriteString("\n" + renderHelp("tab: next field  |  enter: upload  |  esc: cancel"))
	return b.String()
}
