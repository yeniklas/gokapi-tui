package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"
	bubblekey "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
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
	stateList        appState = iota
	stateFilePicker           // upload: pick a file
	stateUploadForm           // upload: configure params
	stateUploading            // upload: spinner
	stateConfirmDel           // uploads: confirm delete
	stateFRCreate             // file requests: create form
	stateFRConfirmDel         // file requests: confirm delete
	stateFRFiles              // file requests: view uploaded files
)

// ── messages ─────────────────────────────────────────────────────────────────

type filesLoadedMsg struct{ files []model.GokapiFile }
type uploadDoneMsg struct{ file model.GokapiFile }
type frLoadedMsg struct{ frs []model.FileRequest }
type frCreatedMsg struct{ fr model.FileRequest }
type frDeletedMsg struct{ frs []model.FileRequest }
type frFilesLoadedMsg struct{ files []model.GokapiFile }
type uploadProgressMsg struct{ pct float64 }
type statusLoadedMsg struct {
	status   model.SystemStatus
	logLines []string
}
type errMsg struct{ err error }
type statusClearMsg struct{}

func (e errMsg) Error() string { return e.err.Error() }

func pressed(msg tea.KeyMsg, b bubblekey.Binding) bool {
	return bubblekey.Matches(msg, b)
}

// ── App ───────────────────────────────────────────────────────────────────────

type App struct {
	// uploads tab
	files    []model.GokapiFile
	cursor   int

	// file requests tab
	fileRequests []model.FileRequest
	frCursor     int
	frInputs     []textinput.Model
	frActiveField int
	frFiles      []model.GokapiFile

	// status tab
	sysStatus model.SystemStatus
	logLines  []string
	logScroll int

	// upload flow
	spinner      spinner.Model
	progressBar  progress.Model
	progressCh   chan float64
	fp           filepicker.Model
	inputs       []textinput.Model
	activeField  int
	selectedPath string

	// shared state
	activeTab int
	state     appState
	statusMsg string
	statusErr bool

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
		cfg:         cfg,
		client:      client,
		spinner:     s,
		progressBar: progress.New(progress.WithDefaultGradient()),
		fp:          fp,
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.loadFilesCmd(), a.loadFRCmd(), loadStatusCmd(a.client), a.spinner.Tick)
}

// ── commands ──────────────────────────────────────────────────────────────────

func (a *App) loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		files, err := a.client.ListFiles(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{files}
	}
}

func (a *App) loadFRCmd() tea.Cmd {
	return func() tea.Msg {
		frs, err := a.client.ListFileRequests(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return frLoadedMsg{frs}
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

func deleteFileCmd(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteFile(context.Background(), id); err != nil {
			return errMsg{err}
		}
		files, err := client.ListFiles(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{files}
	}
}

func createFRCmd(client *api.Client, params model.CreateFileRequestParams) tea.Cmd {
	return func() tea.Msg {
		fr, err := client.CreateFileRequest(context.Background(), params)
		if err != nil {
			return errMsg{err}
		}
		return frCreatedMsg{fr}
	}
}

func deleteFRCmd(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteFileRequest(context.Background(), id); err != nil {
			return errMsg{err}
		}
		frs, err := client.ListFileRequests(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return frDeletedMsg{frs}
	}
}

func loadFRFilesCmd(client *api.Client, frId string) tea.Cmd {
	return func() tea.Msg {
		files, err := client.ListAllFiles(context.Background())
		if err != nil {
			return errMsg{err}
		}
		var frFiles []model.GokapiFile
		for _, f := range files {
			if f.FileRequestId == frId {
				frFiles = append(frFiles, f)
			}
		}
		if frFiles == nil {
			frFiles = []model.GokapiFile{}
		}
		return frFilesLoadedMsg{frFiles}
	}
}

func loadStatusCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := client.GetSystemStatus(context.Background())
		if err != nil {
			return errMsg{err}
		}
		logResp, err := client.GetLogs(context.Background())
		if err != nil {
			return errMsg{err}
		}
		lines := strings.Split(strings.TrimSpace(logResp.LogEntries), "\n")
		if len(lines) == 1 && lines[0] == "" {
			lines = nil
		}
		return statusLoadedMsg{status, lines}
	}
}

func uploadWithProgressCmd(client *api.Client, path string, params model.UploadParams, ch chan float64) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.UploadFileWithProgress(context.Background(), path, params, ch)
		close(ch)
		if err != nil {
			return errMsg{err}
		}
		return uploadDoneMsg{resp.FileInfo}
	}
}

func listenProgressCmd(ch chan float64) tea.Cmd {
	return func() tea.Msg {
		pct, ok := <-ch
		if !ok {
			return nil
		}
		return uploadProgressMsg{pct}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

// ── Update ────────────────────────────────────────────────────────────────────

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.fp.Height = a.height - 6
		a.progressBar.Width = a.width - 4

	case tea.KeyMsg:
		return a.handleKey(msg)

	case filesLoadedMsg:
		a.files = msg.files
		if a.cursor >= len(a.files) && len(a.files) > 0 {
			a.cursor = len(a.files) - 1
		}
		if a.state == stateUploading || a.state == stateConfirmDel {
			a.state = stateList
		}

	case uploadDoneMsg:
		a.state = stateList
		a.files = append(a.files, msg.file)
		a.setStatus("Uploaded: "+msg.file.Name, false)
		return a, tea.Batch(a.loadFilesCmd(), clearStatusAfter(3*time.Second))

	case frLoadedMsg:
		a.fileRequests = msg.frs
		if a.frCursor >= len(a.fileRequests) && len(a.fileRequests) > 0 {
			a.frCursor = len(a.fileRequests) - 1
		}

	case frCreatedMsg:
		a.state = stateList
		a.setStatus("File request created: "+msg.fr.Name, false)
		return a, tea.Batch(a.loadFRCmd(), clearStatusAfter(3*time.Second))

	case frDeletedMsg:
		a.fileRequests = msg.frs
		if a.frCursor >= len(a.fileRequests) && len(a.fileRequests) > 0 {
			a.frCursor = len(a.fileRequests) - 1
		}
		a.state = stateList

	case frFilesLoadedMsg:
		a.frFiles = msg.files

	case statusLoadedMsg:
		a.sysStatus = msg.status
		a.logLines = msg.logLines
		a.logScroll = 0

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

	case uploadProgressMsg:
		cmd := a.progressBar.SetPercent(msg.pct)
		return a, tea.Batch(cmd, listenProgressCmd(a.progressCh))

	case progress.FrameMsg:
		m, cmd := a.progressBar.Update(msg)
		a.progressBar = m.(progress.Model)
		return a, cmd
	}

	// delegate non-key messages to sub-models
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
	case stateFRCreate:
		var cmds []tea.Cmd
		for i := range a.frInputs {
			var cmd tea.Cmd
			a.frInputs[i], cmd = a.frInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	return a, nil
}

// ── key routing ───────────────────────────────────────────────────────────────

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.state {
	case stateFilePicker:
		return a.handleFilePickerKey(msg)
	case stateUploadForm:
		return a.handleUploadFormKey(msg)
	case stateConfirmDel:
		return a.handleConfirmDelKey(msg)
	case stateFRCreate:
		return a.handleFRCreateKey(msg)
	case stateFRConfirmDel:
		return a.handleFRConfirmDelKey(msg)
	case stateFRFiles:
		return a.handleFRFilesKey(msg)
	default:
		switch a.activeTab {
		case tabFileRequests:
			return a.handleFRListKey(msg)
		case tabStatus:
			return a.handleStatusKey(msg)
		default:
			return a.handleListKey(msg)
		}
	}
}

// ── uploads tab ───────────────────────────────────────────────────────────────

func (a *App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Quit):
		return a, tea.Quit
	case pressed(msg, keys.Tab1):
		// already here
	case pressed(msg, keys.Tab2):
		a.activeTab = tabFileRequests
		return a, a.loadFRCmd()
	case pressed(msg, keys.Tab3):
		a.activeTab = tabStatus
		return a, loadStatusCmd(a.client)
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
		if err := clipboard.WriteAll(a.files[a.cursor].UrlDownload); err != nil {
			a.setStatus("Clipboard error: "+err.Error(), true)
			return a, clearStatusAfter(4 * time.Second)
		}
		a.setStatus("Link copied to clipboard", false)
		return a, clearStatusAfter(3 * time.Second)
	case pressed(msg, keys.Delete):
		if len(a.files) > 0 {
			a.state = stateConfirmDel
		}
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
		a.progressCh = make(chan float64, 1)
		cmd := a.progressBar.SetPercent(0)
		return a, tea.Batch(
			cmd,
			uploadWithProgressCmd(a.client, a.selectedPath, params, a.progressCh),
			listenProgressCmd(a.progressCh),
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
		return a, tea.Batch(deleteFileCmd(a.client, id), clearStatusAfter(3*time.Second))
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

// ── file requests tab ─────────────────────────────────────────────────────────

func (a *App) handleFRListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Quit):
		return a, tea.Quit
	case pressed(msg, keys.Tab1):
		a.activeTab = tabUploads
		return a, a.loadFilesCmd()
	case pressed(msg, keys.Tab2):
		// already here
	case pressed(msg, keys.Tab3):
		a.activeTab = tabStatus
		return a, loadStatusCmd(a.client)
	case pressed(msg, keys.Up):
		if a.frCursor > 0 {
			a.frCursor--
		}
	case pressed(msg, keys.Down):
		if a.frCursor < len(a.fileRequests)-1 {
			a.frCursor++
		}
	case pressed(msg, keys.Top):
		a.frCursor = 0
	case pressed(msg, keys.Bottom):
		if len(a.fileRequests) > 0 {
			a.frCursor = len(a.fileRequests) - 1
		}
	case pressed(msg, keys.Refresh):
		return a, a.loadFRCmd()
	case pressed(msg, keys.FRCreate):
		a.frInputs = newFRInputs()
		a.frActiveField = frFieldName
		a.state = stateFRCreate
	case pressed(msg, keys.Delete):
		if len(a.fileRequests) > 0 {
			a.state = stateFRConfirmDel
		}
	case pressed(msg, keys.Yank):
		if len(a.fileRequests) == 0 {
			break
		}
		fr := a.fileRequests[a.frCursor]
		url := a.cfg.ServerURL + "/publicUpload?id=" + fr.Id + "&key=" + fr.ApiKey
		if err := clipboard.WriteAll(url); err != nil {
			a.setStatus("Clipboard error: "+err.Error(), true)
			return a, clearStatusAfter(4 * time.Second)
		}
		a.setStatus("Upload link copied to clipboard", false)
		return a, clearStatusAfter(3 * time.Second)
	case pressed(msg, keys.Confirm):
		if len(a.fileRequests) > 0 {
			a.frFiles = nil
			a.state = stateFRFiles
			return a, loadFRFilesCmd(a.client, a.fileRequests[a.frCursor].Id)
		}
	}
	return a, nil
}

func (a *App) handleFRCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Cancel):
		a.state = stateList
	case msg.String() == "tab":
		a.frInputs[a.frActiveField].Blur()
		a.frActiveField = (a.frActiveField + 1) % frFieldCount
		a.frInputs[a.frActiveField].Focus()
	case pressed(msg, keys.Confirm):
		params := parseFRParams(a.frInputs)
		if params.Name == "" {
			a.setStatus("Name is required", true)
			return a, clearStatusAfter(3 * time.Second)
		}
		a.state = stateUploading
		return a, tea.Batch(createFRCmd(a.client, params), a.spinner.Tick)
	default:
		var cmds []tea.Cmd
		for i := range a.frInputs {
			var cmd tea.Cmd
			a.frInputs[i], cmd = a.frInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}
	return a, nil
}

func (a *App) handleFRConfirmDelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case pressed(msg, keys.Confirm) || msg.String() == "y":
		if len(a.fileRequests) == 0 {
			a.state = stateList
			break
		}
		id := a.fileRequests[a.frCursor].Id
		name := a.fileRequests[a.frCursor].Name
		a.fileRequests = append(a.fileRequests[:a.frCursor], a.fileRequests[a.frCursor+1:]...)
		if a.frCursor >= len(a.fileRequests) && a.frCursor > 0 {
			a.frCursor--
		}
		a.state = stateList
		a.setStatus("Deleted file request: "+name, false)
		return a, tea.Batch(deleteFRCmd(a.client, id), clearStatusAfter(3*time.Second))
	default:
		a.state = stateList
	}
	return a, nil
}

func (a *App) handleFRFilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if pressed(msg, keys.Cancel) || pressed(msg, keys.Quit) {
		a.state = stateList
	}
	return a, nil
}

func (a *App) handleStatusKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxScroll := len(a.logLines) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	switch {
	case pressed(msg, keys.Quit):
		return a, tea.Quit
	case pressed(msg, keys.Tab1):
		a.activeTab = tabUploads
		return a, a.loadFilesCmd()
	case pressed(msg, keys.Tab2):
		a.activeTab = tabFileRequests
		return a, a.loadFRCmd()
	case pressed(msg, keys.Tab3):
		// already here
	case pressed(msg, keys.Refresh):
		return a, loadStatusCmd(a.client)
	case pressed(msg, keys.Up):
		if a.logScroll > 0 {
			a.logScroll--
		}
	case pressed(msg, keys.Down):
		if a.logScroll < maxScroll {
			a.logScroll++
		}
	case pressed(msg, keys.Top):
		a.logScroll = 0
	case pressed(msg, keys.Bottom):
		a.logScroll = maxScroll
	}
	return a, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

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
	case stateFRCreate:
		return a.viewFRCreate()
	case stateFRConfirmDel:
		return a.viewFRConfirmDel()
	case stateFRFiles:
		return a.viewFRFiles()
	default:
		return a.viewList()
	}
}

func (a *App) viewList() string {
	var b strings.Builder
	b.WriteString(renderTabBar(a.activeTab, a.width) + "\n\n")
	switch a.activeTab {
	case tabFileRequests:
		b.WriteString(renderFRList(a.fileRequests, a.frCursor, a.width))
	case tabStatus:
		b.WriteString(renderStatus(a.sysStatus, a.logLines, a.logScroll, a.width, a.height))
	default:
		b.WriteString(renderList(a.files, a.cursor, a.width))
	}
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
	if a.activeTab == tabFileRequests {
		return fmt.Sprintf("\n  %s Creating file request…", a.spinner.View())
	}
	name := filepath.Base(a.selectedPath)
	pct := int(a.progressBar.Percent() * 100)
	return fmt.Sprintf("\n  Uploading %s… %d%%\n\n  %s\n\n  %s",
		name, pct,
		a.progressBar.View(),
		dimStyle.Render("please wait"),
	)
}

func (a *App) viewConfirmDel() string {
	if len(a.files) == 0 {
		return ""
	}
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	return fmt.Sprintf("\n  %s Delete %q ? (y/enter to confirm, any other key to cancel)",
		warn.Render("!"), a.files[a.cursor].Name)
}

func (a *App) viewFRCreate() string {
	return renderFRForm(a.frInputs, a.frActiveField)
}

func (a *App) viewFRConfirmDel() string {
	if len(a.fileRequests) == 0 {
		return ""
	}
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	return fmt.Sprintf("\n  %s Delete file request %q ? (y/enter to confirm, any other key to cancel)",
		warn.Render("!"), a.fileRequests[a.frCursor].Name)
}

func (a *App) viewFRFiles() string {
	if len(a.fileRequests) == 0 {
		return ""
	}
	fr := a.fileRequests[a.frCursor]

	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	b.WriteString(titleStyle.Render(fmt.Sprintf("Files uploaded via: %s", fr.Name)) + "\n\n")
	if a.frFiles == nil {
		b.WriteString(dimStyle.Render("  Loading..."))
	} else {
		b.WriteString(renderList(a.frFiles, -1, a.width))
	}
	b.WriteString("\n" + dimStyle.Render("esc: back"))
	return b.String()
}

func (a *App) statusBar() string {
	var help string
	switch a.activeTab {
	case tabFileRequests:
		help = dimStyle.Render("n:new  y:copy link  d:delete  enter:view files  r:refresh  1/2/3:tabs  q:quit")
	case tabStatus:
		help = dimStyle.Render("j/k:scroll  g/G:top/bottom  r:refresh  1/2/3:tabs  q:quit")
	default:
		help = dimStyle.Render("u:upload  y:copy link  d:delete  r:refresh  1/2/3:tabs  q:quit")
	}
	if a.statusMsg != "" {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		if a.statusErr {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		}
		return style.Render(a.statusMsg) + "  " + help
	}
	switch a.activeTab {
	case tabFileRequests:
		return dimStyle.Render(fmt.Sprintf("%d file request(s)", len(a.fileRequests))) + "  " + help
	case tabStatus:
		return dimStyle.Render(fmt.Sprintf("%d log line(s)", len(a.logLines))) + "  " + help
	default:
		return dimStyle.Render(fmt.Sprintf("%d file(s)", len(a.files))) + "  " + help
	}
}

func (a *App) setStatus(msg string, isErr bool) {
	a.statusMsg = msg
	a.statusErr = isErr
}
