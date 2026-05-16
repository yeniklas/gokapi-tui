package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/filepicker"
	bubblekey "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yeniklas/gokapi-tui/internal/api"
	"github.com/yeniklas/gokapi-tui/internal/config"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

type appState int

const (
	stateList appState = iota
	stateFilePicker
	stateUploadForm
	stateUploading
	stateConfirmDel
)

// messages
type filesLoadedMsg struct{ files []model.GokapiFile }
type uploadDoneMsg struct{ file model.GokapiFile }
type errMsg struct{ err error }
type statusClearMsg struct{}

func (e errMsg) Error() string { return e.err.Error() }

func pressed(msg tea.KeyMsg, b bubblekey.Binding) bool {
	return bubblekey.Matches(msg, b)
}

type App struct {
	files     []model.GokapiFile
	cursor    int
	state     appState
	statusMsg string
	statusErr bool

	spinner    spinner.Model
	fp         filepicker.Model
	inputs     []textinput.Model
	activeField int
	selectedPath string

	cfg    *config.Config
	client *api.Client
	width  int
	height int
}

func New(cfg *config.Config, client *api.Client) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	fp := filepicker.New()
	fp.ShowHidden = false

	return &App{
		cfg:     cfg,
		client:  client,
		spinner: s,
		fp:      fp,
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.loadFilesCmd(), a.spinner.Tick)
}

func (a *App) loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		files, err := a.client.ListFiles(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{files}
	}
}

func uploadCmd(client *api.Client, path string, params model.UploadParams) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.UploadFile(context.Background(), path, params)
		if err != nil {
			return errMsg{err}
		}
		return uploadDoneMsg{resp.FileInfo}
	}
}

func deleteCmd(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteFile(context.Background(), id); err != nil {
			return errMsg{err}
		}
		files, err := client.ListFiles(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{files: files}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.fp.Height = a.height - 6

	case tea.KeyMsg:
		return a.handleKey(msg)

	case filesLoadedMsg:
		a.files = msg.files
		a.state = stateList
		if a.cursor >= len(a.files) && len(a.files) > 0 {
			a.cursor = len(a.files) - 1
		}

	case uploadDoneMsg:
		a.state = stateList
		a.files = append(a.files, msg.file)
		a.setStatus("Uploaded: "+msg.file.Name, false)
		return a, tea.Batch(a.loadFilesCmd(), clearStatusAfter(3*time.Second))

	case errMsg:
		a.state = stateList
		a.setStatus("Error: "+msg.err.Error(), true)
		return a, clearStatusAfter(5 * time.Second)

	case statusClearMsg:
		a.statusMsg = ""
		a.statusErr = false

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	}

	// delegate to sub-models when needed (non-key messages)
	switch a.state {
	case stateFilePicker:
		var cmd tea.Cmd
		a.fp, cmd = a.fp.Update(msg)
		return a, cmd
	case stateUploadForm:
		var cmds []tea.Cmd
		for i := range a.inputs {
			var cmd tea.Cmd
			a.inputs[i], cmd = a.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.state {
	case stateFilePicker:
		return a.handleFilePickerKey(msg)
	case stateUploadForm:
		return a.handleUploadFormKey(msg)
	case stateConfirmDel:
		return a.handleConfirmDelKey(msg)
	default:
		return a.handleListKey(msg)
	}
}

func (a *App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Quit):
		return a, tea.Quit

	case pressed(msg, keys.Up):
		if a.cursor > 0 {
			a.cursor--
		}

	case pressed(msg, keys.Down):
		if a.cursor < len(a.files)-1 {
			a.cursor++
		}

	case pressed(msg, keys.Top):
		a.cursor = 0

	case pressed(msg, keys.Bottom):
		if len(a.files) > 0 {
			a.cursor = len(a.files) - 1
		}

	case pressed(msg, keys.Refresh):
		return a, a.loadFilesCmd()

	case pressed(msg, keys.Upload):
		a.fp = filepicker.New()
		a.fp.ShowHidden = false
		a.fp.Height = a.height - 6
		a.state = stateFilePicker
		return a, a.fp.Init()

	case pressed(msg, keys.Yank):
		if len(a.files) == 0 {
			break
		}
		url := a.files[a.cursor].UrlDownload
		if err := clipboard.WriteAll(url); err != nil {
			a.setStatus("Clipboard error: "+err.Error(), true)
			return a, clearStatusAfter(4 * time.Second)
		}
		a.setStatus("Link copied to clipboard", false)
		return a, clearStatusAfter(3 * time.Second)

	case pressed(msg, keys.Delete):
		if len(a.files) == 0 {
			break
		}
		a.state = stateConfirmDel
	}
	return a, nil
}

func (a *App) handleFilePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if pressed(msg, keys.Cancel) {
		a.state = stateList
		return a, nil
	}
	var cmd tea.Cmd
	a.fp, cmd = a.fp.Update(msg)
	if selected, path := a.fp.DidSelectFile(msg); selected {
		a.selectedPath = path
		a.inputs = newUploadInputs(a.cfg.Defaults.AllowedDownloads, a.cfg.Defaults.ExpiryDays, a.cfg.Defaults.Password)
		a.activeField = fieldDownloads
		a.state = stateUploadForm
	} else if disabled, _ := a.fp.DidSelectDisabledFile(msg); disabled {
		a.setStatus("Cannot select that file", true)
		return a, tea.Batch(cmd, clearStatusAfter(3*time.Second))
	}
	return a, cmd
}

func (a *App) handleUploadFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Cancel):
		a.state = stateList

	case msg.String() == "tab":
		a.inputs[a.activeField].Blur()
		a.activeField = (a.activeField + 1) % fieldCount
		a.inputs[a.activeField].Focus()

	case pressed(msg, keys.Confirm):
		params := a.parseUploadParams()
		a.state = stateUploading
		return a, tea.Batch(
			uploadCmd(a.client, a.selectedPath, params),
			a.spinner.Tick,
		)

	default:
		var cmds []tea.Cmd
		for i := range a.inputs {
			var cmd tea.Cmd
			a.inputs[i], cmd = a.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}
	return a, nil
}

func (a *App) handleConfirmDelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Confirm) || msg.String() == "y":
		if len(a.files) == 0 {
			a.state = stateList
			break
		}
		id := a.files[a.cursor].Id
		name := a.files[a.cursor].Name
		a.files = append(a.files[:a.cursor], a.files[a.cursor+1:]...)
		if a.cursor >= len(a.files) && a.cursor > 0 {
			a.cursor--
		}
		a.state = stateList
		a.setStatus("Deleted: "+name, false)
		return a, tea.Batch(deleteCmd(a.client, id), clearStatusAfter(3*time.Second))
	default:
		a.state = stateList
	}
	return a, nil
}

func (a *App) parseUploadParams() model.UploadParams {
	p := model.UploadParams{
		AllowedDownloads: a.cfg.Defaults.AllowedDownloads,
		ExpiryDays:       a.cfg.Defaults.ExpiryDays,
		Password:         a.cfg.Defaults.Password,
	}
	if v, err := strconv.Atoi(strings.TrimSpace(a.inputs[fieldDownloads].Value())); err == nil && v > 0 {
		p.AllowedDownloads = v
	}
	if v, err := strconv.Atoi(strings.TrimSpace(a.inputs[fieldExpiry].Value())); err == nil && v > 0 {
		p.ExpiryDays = v
	}
	p.Password = a.inputs[fieldPassword].Value()
	return p
}

func (a *App) View() string {
	switch a.state {
	case stateFilePicker:
		return a.viewFilePicker()
	case stateUploadForm:
		return a.viewUploadForm()
	case stateUploading:
		return a.viewUploading()
	case stateConfirmDel:
		return a.viewConfirmDel()
	default:
		return a.viewList()
	}
}

func (a *App) viewList() string {
	var b strings.Builder
	b.WriteString(renderList(a.files, a.cursor, a.width))
	b.WriteString("\n")
	b.WriteString(a.statusBar())
	return b.String()
}

func (a *App) viewFilePicker() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	b.WriteString(titleStyle.Render("Select a file to upload") + "\n\n")
	b.WriteString(a.fp.View())
	b.WriteString("\n" + dimStyle.Render("esc: cancel"))
	return b.String()
}

func (a *App) viewUploadForm() string {
	return renderUploadForm(a.inputs, a.activeField, a.selectedPath)
}

func (a *App) viewUploading() string {
	return fmt.Sprintf("\n  %s Uploading %s...", a.spinner.View(), a.selectedPath)
}

func (a *App) viewConfirmDel() string {
	if len(a.files) == 0 {
		return ""
	}
	name := a.files[a.cursor].Name
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	return fmt.Sprintf("\n  %s Delete %q ? (y/enter to confirm, any other key to cancel)",
		warn.Render("!"), name)
}

func (a *App) statusBar() string {
	help := dimStyle.Render("u:upload  y:copy link  d:delete  r:refresh  q:quit")
	if a.statusMsg != "" {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		if a.statusErr {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		}
		return style.Render(a.statusMsg) + "  " + help
	}
	fileCount := dimStyle.Render(fmt.Sprintf("%d file(s)", len(a.files)))
	return fileCount + "  " + help
}

func (a *App) setStatus(msg string, isErr bool) {
	a.statusMsg = msg
	a.statusErr = isErr
}

