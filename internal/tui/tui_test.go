package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yeniklas/gokapi-tui/internal/config"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func testApp(files []model.GokapiFile) *App {
	cfg := &config.Config{
		ServerURL: "http://test",
		APIKey:    "testkey",
		Defaults: config.Defaults{
			AllowedDownloads: 3,
			ExpiryDays:       14,
		},
	}
	app := New(cfg, nil)
	app.files = files
	app.width = 120
	app.height = 40
	return app
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func updateApp(app *App, key string) *App {
	m, _ := app.Update(keyMsg(key))
	return m.(*App)
}

func sampleFiles() []model.GokapiFile {
	return []model.GokapiFile{
		{Id: "a1", Name: "alpha.pdf", Size: "1 MB", UrlDownload: "http://srv/d/a1"},
		{Id: "b2", Name: "beta.zip", Size: "2 MB", UrlDownload: "http://srv/d/b2"},
		{Id: "c3", Name: "gamma.txt", Size: "3 KB", UrlDownload: "http://srv/d/c3"},
	}
}

// ── truncate ─────────────────────────────────────────────────────────────────

func TestTruncate_ShortString(t *testing.T) {
	if got := truncate("hi", 10); got != "hi" {
		t.Errorf("got %q, want %q", got, "hi")
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	if got := truncate("hello", 5); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestTruncate_LongString(t *testing.T) {
	got := truncate("abcdefghij", 7)
	if got != "abcd..." {
		t.Errorf("got %q, want %q", got, "abcd...")
	}
}

func TestTruncate_TinyMax(t *testing.T) {
	got := truncate("hello", 2)
	if len(got) != 2 {
		t.Errorf("got len %d, want 2", len(got))
	}
}

// ── renderList ────────────────────────────────────────────────────────────────

func TestRenderList_ContainsFileNames(t *testing.T) {
	files := sampleFiles()
	out := renderList(files, 0, 120)
	for _, f := range files {
		if !strings.Contains(out, f.Name) {
			t.Errorf("output missing file name %q", f.Name)
		}
	}
}

func TestRenderList_CursorRow(t *testing.T) {
	files := sampleFiles()
	out := renderList(files, 1, 120)
	lines := strings.Split(out, "\n")
	// line 0 = header, line 1 = file[0], line 2 = file[1] (cursor)
	if !strings.Contains(lines[2], ">") {
		t.Errorf("cursor row should contain '>', got: %q", lines[2])
	}
}

func TestRenderList_EmptyFiles(t *testing.T) {
	out := renderList(nil, 0, 120)
	if !strings.Contains(out, "No files found") {
		t.Errorf("expected empty-state message, got: %q", out)
	}
}

func TestRenderList_UnlimitedTime(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", UnlimitedTime: true}}
	out := renderList(files, 0, 120)
	if !strings.Contains(out, "never") {
		t.Errorf("unlimited time should show 'never', got: %q", out)
	}
}

func TestRenderList_UnlimitedDownloads(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", UnlimitedDownloads: true}}
	out := renderList(files, 0, 120)
	if !strings.Contains(out, "unlimited") {
		t.Errorf("unlimited downloads should show 'unlimited', got: %q", out)
	}
}

func TestRenderList_UploadDate(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", UploadDate: 1700000000}}
	out := renderList(files, 0, 120)
	if !strings.Contains(out, "2023-11-14") {
		t.Errorf("expected formatted upload date in output, got: %q", out)
	}
}

func TestRenderList_NoUploadDate(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", UploadDate: 0}}
	out := renderList(files, 0, 120)
	// should not panic and should still render the row
	if !strings.Contains(out, "f.txt") {
		t.Errorf("row missing when UploadDate is zero, got: %q", out)
	}
}

func TestRenderList_PasswordProtected(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", IsPasswordProtected: true}}
	out := renderList(files, 0, 120)
	if !strings.Contains(out, "yes") {
		t.Errorf("password protected file should show 'yes', got: %q", out)
	}
}

func TestRenderList_NotPasswordProtected(t *testing.T) {
	files := []model.GokapiFile{{Id: "x", Name: "f.txt", IsPasswordProtected: false}}
	out := renderList(files, 0, 120)
	if strings.Contains(out, "yes") {
		t.Errorf("non-password-protected file should not show 'yes', got: %q", out)
	}
}

// ── newUploadInputs ───────────────────────────────────────────────────────────

func TestNewUploadInputs_Values(t *testing.T) {
	inputs := newUploadInputs(5, 21, "mypass")
	if inputs[fieldDownloads].Value() != "5" {
		t.Errorf("downloads input = %q, want %q", inputs[fieldDownloads].Value(), "5")
	}
	if inputs[fieldExpiry].Value() != "21" {
		t.Errorf("expiry input = %q, want %q", inputs[fieldExpiry].Value(), "21")
	}
	if inputs[fieldPassword].Value() != "mypass" {
		t.Errorf("password input = %q, want %q", inputs[fieldPassword].Value(), "mypass")
	}
}

func TestNewUploadInputs_FirstFieldFocused(t *testing.T) {
	inputs := newUploadInputs(1, 7, "")
	if !inputs[fieldDownloads].Focused() {
		t.Error("first field (downloads) should be focused")
	}
	if inputs[fieldExpiry].Focused() {
		t.Error("expiry field should not be focused")
	}
}

// ── parseUploadParams ─────────────────────────────────────────────────────────

func TestParseUploadParams_ValidIntegers(t *testing.T) {
	app := testApp(nil)
	app.inputs = newUploadInputs(1, 7, "")
	app.inputs[fieldDownloads].SetValue("10")
	app.inputs[fieldExpiry].SetValue("30")
	app.inputs[fieldPassword].SetValue("pw123")

	p := app.parseUploadParams()
	if p.AllowedDownloads != 10 {
		t.Errorf("AllowedDownloads = %d, want 10", p.AllowedDownloads)
	}
	if p.ExpiryDays != 30 {
		t.Errorf("ExpiryDays = %d, want 30", p.ExpiryDays)
	}
	if p.Password != "pw123" {
		t.Errorf("Password = %q, want %q", p.Password, "pw123")
	}
}

func TestParseUploadParams_InvalidFallsToDefaults(t *testing.T) {
	app := testApp(nil)
	app.inputs = newUploadInputs(1, 7, "")
	app.inputs[fieldDownloads].SetValue("not-a-number")
	app.inputs[fieldExpiry].SetValue("")

	p := app.parseUploadParams()
	if p.AllowedDownloads != app.cfg.Defaults.AllowedDownloads {
		t.Errorf("AllowedDownloads = %d, want default %d", p.AllowedDownloads, app.cfg.Defaults.AllowedDownloads)
	}
	if p.ExpiryDays != app.cfg.Defaults.ExpiryDays {
		t.Errorf("ExpiryDays = %d, want default %d", p.ExpiryDays, app.cfg.Defaults.ExpiryDays)
	}
}

func TestParseUploadParams_ZeroFallsToDefaults(t *testing.T) {
	app := testApp(nil)
	app.inputs = newUploadInputs(1, 7, "")
	app.inputs[fieldDownloads].SetValue("0")
	app.inputs[fieldExpiry].SetValue("0")

	p := app.parseUploadParams()
	if p.AllowedDownloads != app.cfg.Defaults.AllowedDownloads {
		t.Errorf("zero AllowedDownloads should fall back to default %d, got %d",
			app.cfg.Defaults.AllowedDownloads, p.AllowedDownloads)
	}
}

// ── App state machine ─────────────────────────────────────────────────────────

func TestUpdate_FilesLoadedMsg(t *testing.T) {
	app := testApp(nil)
	files := sampleFiles()
	m, _ := app.Update(filesLoadedMsg{files: files})
	got := m.(*App)
	if len(got.files) != len(files) {
		t.Errorf("got %d files, want %d", len(got.files), len(files))
	}
	if got.state != stateList {
		t.Errorf("state = %d, want stateList", got.state)
	}
}

func TestUpdate_FilesLoadedMsg_ClampsCursor(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 2
	// reload with fewer files
	m, _ := app.Update(filesLoadedMsg{files: sampleFiles()[:1]})
	got := m.(*App)
	if got.cursor != 0 {
		t.Errorf("cursor should clamp to 0, got %d", got.cursor)
	}
}

func TestUpdate_UploadDoneMsg(t *testing.T) {
	app := testApp(sampleFiles())
	newFile := model.GokapiFile{Id: "new", Name: "new.txt"}
	m, _ := app.Update(uploadDoneMsg{file: newFile})
	got := m.(*App)
	if got.state != stateList {
		t.Errorf("state = %d, want stateList", got.state)
	}
	if got.statusMsg == "" {
		t.Error("expected status message after upload")
	}
	if got.statusErr {
		t.Error("upload success should not set statusErr")
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	app := testApp(nil)
	app.state = stateUploading
	m, _ := app.Update(errMsg{err: &testError{"something failed"}})
	got := m.(*App)
	if got.state != stateList {
		t.Errorf("state = %d, want stateList", got.state)
	}
	if !got.statusErr {
		t.Error("errMsg should set statusErr = true")
	}
	if !strings.Contains(got.statusMsg, "something failed") {
		t.Errorf("statusMsg = %q should contain error text", got.statusMsg)
	}
}

func TestUpdate_StatusClearMsg(t *testing.T) {
	app := testApp(nil)
	app.statusMsg = "some message"
	app.statusErr = true
	m, _ := app.Update(statusClearMsg{})
	got := m.(*App)
	if got.statusMsg != "" {
		t.Errorf("statusMsg should be cleared, got %q", got.statusMsg)
	}
	if got.statusErr {
		t.Error("statusErr should be cleared")
	}
}

func TestHandleListKey_Down(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 0
	app = updateApp(app, "j")
	if app.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", app.cursor)
	}
}

func TestHandleListKey_DownClampAtEnd(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 2
	app = updateApp(app, "j")
	if app.cursor != 2 {
		t.Errorf("cursor should stay at 2 at boundary, got %d", app.cursor)
	}
}

func TestHandleListKey_Up(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 2
	app = updateApp(app, "k")
	if app.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after k", app.cursor)
	}
}

func TestHandleListKey_UpClampAtStart(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 0
	app = updateApp(app, "k")
	if app.cursor != 0 {
		t.Errorf("cursor should stay at 0 at boundary, got %d", app.cursor)
	}
}

func TestHandleListKey_JumpTop(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 2
	app = updateApp(app, "g")
	if app.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after g", app.cursor)
	}
}

func TestHandleListKey_JumpBottom(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 0
	app = updateApp(app, "G")
	if app.cursor != 2 {
		t.Errorf("cursor = %d, want 2 after G", app.cursor)
	}
}

func TestHandleListKey_DeleteTransitionsState(t *testing.T) {
	app := testApp(sampleFiles())
	app = updateApp(app, "d")
	if app.state != stateConfirmDel {
		t.Errorf("state = %d, want stateConfirmDel after d", app.state)
	}
}

func TestHandleListKey_DeleteOnEmptyDoesNothing(t *testing.T) {
	app := testApp(nil)
	app = updateApp(app, "d")
	if app.state != stateList {
		t.Errorf("delete on empty list should stay in stateList, got %d", app.state)
	}
}

func TestHandleConfirmDelKey_ConfirmRemovesFile(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 1
	app.state = stateConfirmDel
	app = updateApp(app, "y")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after confirm delete", app.state)
	}
	if len(app.files) != 2 {
		t.Errorf("files len = %d, want 2 after delete", len(app.files))
	}
	if app.files[0].Id != "a1" || app.files[1].Id != "c3" {
		t.Errorf("wrong files remaining: %+v", app.files)
	}
	if app.statusMsg == "" {
		t.Error("expected status message after delete")
	}
}

func TestHandleConfirmDelKey_EnterConfirmsDelete(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 0
	app.state = stateConfirmDel
	app = updateApp(app, "enter")
	if len(app.files) != 2 {
		t.Errorf("files len = %d, want 2 after confirm delete", len(app.files))
	}
}

func TestHandleConfirmDelKey_CancelRestoresList(t *testing.T) {
	app := testApp(sampleFiles())
	app.state = stateConfirmDel
	app = updateApp(app, "n")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after cancel", app.state)
	}
	if len(app.files) != 3 {
		t.Errorf("files should be unchanged after cancel, got %d", len(app.files))
	}
}

func TestHandleConfirmDelKey_DeleteLastFile_ClampsCursor(t *testing.T) {
	app := testApp(sampleFiles()[:1])
	app.cursor = 0
	app.state = stateConfirmDel
	app = updateApp(app, "y")
	if app.cursor != 0 {
		t.Errorf("cursor should remain 0 after deleting only file, got %d", app.cursor)
	}
}

func TestHandleConfirmDelKey_DeleteLastClampsToNewEnd(t *testing.T) {
	app := testApp(sampleFiles())
	app.cursor = 2 // last item
	app.state = stateConfirmDel
	app = updateApp(app, "y")
	if app.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after deleting last item", app.cursor)
	}
}

func TestHandleUploadFormKey_EscReturnsToList(t *testing.T) {
	app := testApp(nil)
	app.state = stateUploadForm
	app.inputs = newUploadInputs(1, 7, "")
	app = updateApp(app, "esc")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after esc", app.state)
	}
}

func TestHandleUploadFormKey_TabCyclesField(t *testing.T) {
	app := testApp(nil)
	app.state = stateUploadForm
	app.inputs = newUploadInputs(1, 7, "")
	app.activeField = fieldDownloads

	app = updateApp(app, "tab")
	if app.activeField != fieldExpiry {
		t.Errorf("activeField = %d, want fieldExpiry (%d)", app.activeField, fieldExpiry)
	}
	app = updateApp(app, "tab")
	if app.activeField != fieldPassword {
		t.Errorf("activeField = %d, want fieldPassword (%d)", app.activeField, fieldPassword)
	}
	app = updateApp(app, "tab")
	if app.activeField != fieldDownloads {
		t.Errorf("activeField should wrap to fieldDownloads (%d), got %d", fieldDownloads, app.activeField)
	}
}

func TestHandleUploadFormKey_EnterStartsUpload(t *testing.T) {
	app := testApp(nil)
	app.state = stateUploadForm
	app.inputs = newUploadInputs(1, 7, "")
	app.selectedPath = "/tmp/somefile.txt"
	app = updateApp(app, "enter")
	if app.state != stateUploading {
		t.Errorf("state = %d, want stateUploading after enter", app.state)
	}
}

// ── tab switching ─────────────────────────────────────────────────────────────

func TestTabSwitch_ToFileRequests(t *testing.T) {
	app := testApp(sampleFiles())
	app = updateApp(app, "2")
	if app.activeTab != tabFileRequests {
		t.Errorf("activeTab = %d, want tabFileRequests (%d)", app.activeTab, tabFileRequests)
	}
}

func TestTabSwitch_BackToUploads(t *testing.T) {
	app := testApp(sampleFiles())
	app.activeTab = tabFileRequests
	app = updateApp(app, "1")
	if app.activeTab != tabUploads {
		t.Errorf("activeTab = %d, want tabUploads (%d)", app.activeTab, tabUploads)
	}
}

// ── renderTabBar ──────────────────────────────────────────────────────────────

func TestRenderTabBar_ActiveBracketed(t *testing.T) {
	out := renderTabBar(tabUploads, 120)
	if !strings.Contains(out, "[Uploads]") {
		t.Errorf("active tab should be bracketed, got: %q", out)
	}
}

func TestRenderTabBar_InactiveNoBrackets(t *testing.T) {
	out := renderTabBar(tabUploads, 120)
	if strings.Contains(out, "[File Requests]") {
		t.Errorf("inactive tab should not be bracketed, got: %q", out)
	}
	if !strings.Contains(out, "File Requests") {
		t.Errorf("inactive tab name should still appear, got: %q", out)
	}
}

// ── renderFRList ──────────────────────────────────────────────────────────────

func sampleFRs() []model.FileRequest {
	return []model.FileRequest{
		{Id: "r1", Name: "Photos", MaxFiles: 10, UploadedFiles: 3},
		{Id: "r2", Name: "Documents", MaxFiles: 0, UploadedFiles: 0},
	}
}

func TestRenderFRList_ContainsNames(t *testing.T) {
	frs := sampleFRs()
	out := renderFRList(frs, 0, 120)
	for _, fr := range frs {
		if !strings.Contains(out, fr.Name) {
			t.Errorf("output missing FR name %q", fr.Name)
		}
	}
}

func TestRenderFRList_CursorRow(t *testing.T) {
	frs := sampleFRs()
	out := renderFRList(frs, 1, 120)
	lines := strings.Split(out, "\n")
	// line 0 = header, line 1 = fr[0], line 2 = fr[1] (cursor)
	if !strings.Contains(lines[2], ">") {
		t.Errorf("cursor row should contain '>', got: %q", lines[2])
	}
}

func TestRenderFRList_EmptyState(t *testing.T) {
	out := renderFRList(nil, 0, 120)
	if !strings.Contains(out, "No file requests found") {
		t.Errorf("expected empty-state message, got: %q", out)
	}
}

func TestRenderFRList_UnlimitedMaxFiles(t *testing.T) {
	frs := []model.FileRequest{{Id: "x", Name: "r", MaxFiles: 0}}
	out := renderFRList(frs, 0, 120)
	if !strings.Contains(out, "unlimited") {
		t.Errorf("zero MaxFiles should show 'unlimited', got: %q", out)
	}
}

func TestRenderFRList_NoExpiry(t *testing.T) {
	frs := []model.FileRequest{{Id: "x", Name: "r", Expiry: 0}}
	out := renderFRList(frs, 0, 120)
	if !strings.Contains(out, "never") {
		t.Errorf("zero Expiry should show 'never', got: %q", out)
	}
}

func TestRenderFRList_ExpiryDate(t *testing.T) {
	frs := []model.FileRequest{{Id: "x", Name: "r", Expiry: 1700000000}}
	out := renderFRList(frs, 0, 120)
	if !strings.Contains(out, "2023-11-14") {
		t.Errorf("expected formatted expiry date, got: %q", out)
	}
}

// ── formatBytes ───────────────────────────────────────────────────────────────

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1 << 20, "1.0 MB"},
		{1 << 30, "1.0 GB"},
	}
	for _, tc := range tests {
		if got := formatBytes(tc.input); got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── parseInt ──────────────────────────────────────────────────────────────────

func TestParseInt(t *testing.T) {
	if parseInt("42") != 42 {
		t.Error("parseInt(42) should return 42")
	}
	if parseInt("") != 0 {
		t.Error("parseInt(\"\") should return 0")
	}
	if parseInt("abc") != 0 {
		t.Error("parseInt(abc) should return 0")
	}
	if parseInt("  7  ") != 7 {
		t.Error("parseInt with whitespace should return 7")
	}
}

// ── FR message handling ───────────────────────────────────────────────────────

func TestUpdate_FRLoadedMsg(t *testing.T) {
	app := testApp(nil)
	frs := sampleFRs()
	m, _ := app.Update(frLoadedMsg{frs: frs})
	got := m.(*App)
	if len(got.fileRequests) != len(frs) {
		t.Errorf("got %d file requests, want %d", len(got.fileRequests), len(frs))
	}
}

func TestUpdate_FRLoadedMsg_ClampsCursor(t *testing.T) {
	app := testApp(nil)
	app.fileRequests = sampleFRs()
	app.frCursor = 1
	m, _ := app.Update(frLoadedMsg{frs: sampleFRs()[:1]})
	got := m.(*App)
	if got.frCursor != 0 {
		t.Errorf("frCursor should clamp to 0, got %d", got.frCursor)
	}
}

func TestUpdate_FRCreatedMsg(t *testing.T) {
	app := testApp(nil)
	app.activeTab = tabFileRequests
	app.state = stateUploading
	fr := model.FileRequest{Id: "new", Name: "New Request"}
	m, _ := app.Update(frCreatedMsg{fr: fr})
	got := m.(*App)
	if got.state != stateList {
		t.Errorf("state = %d, want stateList after frCreatedMsg", got.state)
	}
	if got.statusMsg == "" {
		t.Error("expected status message after FR creation")
	}
}

func TestUpdate_FRDeletedMsg(t *testing.T) {
	app := testApp(nil)
	app.fileRequests = sampleFRs()
	app.frCursor = 1
	remaining := sampleFRs()[:1]
	m, _ := app.Update(frDeletedMsg{frs: remaining})
	got := m.(*App)
	if len(got.fileRequests) != 1 {
		t.Errorf("got %d file requests, want 1", len(got.fileRequests))
	}
	if got.frCursor != 0 {
		t.Errorf("frCursor should clamp to 0, got %d", got.frCursor)
	}
}

// ── FR list navigation ────────────────────────────────────────────────────────

func frApp() *App {
	app := testApp(sampleFiles())
	app.activeTab = tabFileRequests
	app.fileRequests = sampleFRs()
	return app
}

func TestHandleFRListKey_Down(t *testing.T) {
	app := frApp()
	app.frCursor = 0
	app = updateApp(app, "j")
	if app.frCursor != 1 {
		t.Errorf("frCursor = %d, want 1 after j", app.frCursor)
	}
}

func TestHandleFRListKey_DownClampAtEnd(t *testing.T) {
	app := frApp()
	app.frCursor = 1
	app = updateApp(app, "j")
	if app.frCursor != 1 {
		t.Errorf("frCursor should stay at 1 at boundary, got %d", app.frCursor)
	}
}

func TestHandleFRListKey_Up(t *testing.T) {
	app := frApp()
	app.frCursor = 1
	app = updateApp(app, "k")
	if app.frCursor != 0 {
		t.Errorf("frCursor = %d, want 0 after k", app.frCursor)
	}
}

func TestHandleFRListKey_NewOpensCreateForm(t *testing.T) {
	app := frApp()
	app = updateApp(app, "n")
	if app.state != stateFRCreate {
		t.Errorf("state = %d, want stateFRCreate after n", app.state)
	}
}

func TestHandleFRListKey_DeleteOpensConfirm(t *testing.T) {
	app := frApp()
	app = updateApp(app, "d")
	if app.state != stateFRConfirmDel {
		t.Errorf("state = %d, want stateFRConfirmDel after d", app.state)
	}
}

func TestHandleFRListKey_DeleteOnEmptyDoesNothing(t *testing.T) {
	app := frApp()
	app.fileRequests = nil
	app = updateApp(app, "d")
	if app.state != stateList {
		t.Errorf("delete on empty FR list should stay in stateList, got %d", app.state)
	}
}

func TestHandleFRListKey_EnterViewsFiles(t *testing.T) {
	app := frApp()
	app = updateApp(app, "enter")
	if app.state != stateFRFiles {
		t.Errorf("state = %d, want stateFRFiles after enter", app.state)
	}
}

func TestHandleFRListKey_EnterOnEmptyDoesNothing(t *testing.T) {
	app := frApp()
	app.fileRequests = nil
	app = updateApp(app, "enter")
	if app.state != stateList {
		t.Errorf("enter on empty FR list should stay in stateList, got %d", app.state)
	}
}

// ── FR confirm delete ─────────────────────────────────────────────────────────

func TestHandleFRConfirmDelKey_ConfirmRemovesFR(t *testing.T) {
	app := frApp()
	app.frCursor = 0
	app.state = stateFRConfirmDel
	app = updateApp(app, "y")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after confirm", app.state)
	}
	if len(app.fileRequests) != 1 {
		t.Errorf("fileRequests len = %d, want 1 after delete", len(app.fileRequests))
	}
	if app.fileRequests[0].Id != "r2" {
		t.Errorf("wrong FR remaining: %+v", app.fileRequests)
	}
}

func TestHandleFRConfirmDelKey_CancelKeepsFRs(t *testing.T) {
	app := frApp()
	app.state = stateFRConfirmDel
	app = updateApp(app, "esc")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after cancel", app.state)
	}
	if len(app.fileRequests) != 2 {
		t.Errorf("fileRequests should be unchanged after cancel, got %d", len(app.fileRequests))
	}
}

// ── FR create form ────────────────────────────────────────────────────────────

func TestHandleFRCreateKey_EscReturnsToList(t *testing.T) {
	app := frApp()
	app.state = stateFRCreate
	app.frInputs = newFRInputs()
	app = updateApp(app, "esc")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after esc", app.state)
	}
}

func TestHandleFRCreateKey_TabCyclesField(t *testing.T) {
	app := frApp()
	app.state = stateFRCreate
	app.frInputs = newFRInputs()
	app.frActiveField = frFieldName

	app = updateApp(app, "tab")
	if app.frActiveField != frFieldNotes {
		t.Errorf("frActiveField = %d, want frFieldNotes (%d)", app.frActiveField, frFieldNotes)
	}
	app = updateApp(app, "tab")
	if app.frActiveField != frFieldExpiry {
		t.Errorf("frActiveField = %d, want frFieldExpiry (%d)", app.frActiveField, frFieldExpiry)
	}
}

func TestHandleFRCreateKey_EnterWithNameStartsCreation(t *testing.T) {
	app := frApp()
	app.state = stateFRCreate
	app.frInputs = newFRInputs()
	app.frInputs[frFieldName].SetValue("My Request")
	app = updateApp(app, "enter")
	if app.state != stateUploading {
		t.Errorf("state = %d, want stateUploading after enter with name", app.state)
	}
}

func TestHandleFRCreateKey_EnterWithoutNameShowsError(t *testing.T) {
	app := frApp()
	app.state = stateFRCreate
	app.frInputs = newFRInputs()
	// name is empty by default
	app = updateApp(app, "enter")
	if app.state == stateUploading {
		t.Error("should not start creation without a name")
	}
	if !app.statusErr {
		t.Error("should set error status when name is missing")
	}
}

// ── parseFRParams ─────────────────────────────────────────────────────────────

func TestParseFRParams_BasicFields(t *testing.T) {
	inputs := newFRInputs()
	inputs[frFieldName].SetValue("Test Request")
	inputs[frFieldNotes].SetValue("some notes")
	inputs[frFieldMaxFiles].SetValue("5")
	inputs[frFieldMaxSize].SetValue("100")

	p := parseFRParams(inputs)
	if p.Name != "Test Request" {
		t.Errorf("Name = %q, want %q", p.Name, "Test Request")
	}
	if p.Notes != "some notes" {
		t.Errorf("Notes = %q, want %q", p.Notes, "some notes")
	}
	if p.MaxFiles != 5 {
		t.Errorf("MaxFiles = %d, want 5", p.MaxFiles)
	}
	if p.MaxSize != 100 {
		t.Errorf("MaxSize = %d, want 100", p.MaxSize)
	}
}

func TestParseFRParams_ZeroExpiryMeansNever(t *testing.T) {
	inputs := newFRInputs()
	inputs[frFieldExpiry].SetValue("0")
	p := parseFRParams(inputs)
	if p.ExpiryAt != 0 {
		t.Errorf("ExpiryAt should be 0 for zero days, got %d", p.ExpiryAt)
	}
}

func TestParseFRParams_PositiveExpirySetsFutureTimestamp(t *testing.T) {
	inputs := newFRInputs()
	inputs[frFieldExpiry].SetValue("7")
	p := parseFRParams(inputs)
	if p.ExpiryAt <= 0 {
		t.Error("ExpiryAt should be a future unix timestamp for positive days")
	}
}

// ── FR files view ─────────────────────────────────────────────────────────────

func TestHandleFRFilesKey_EscReturnsToList(t *testing.T) {
	app := frApp()
	app.state = stateFRFiles
	app = updateApp(app, "esc")
	if app.state != stateList {
		t.Errorf("state = %d, want stateList after esc", app.state)
	}
}

// testError implements the error interface for tests
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
