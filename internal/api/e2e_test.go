//go:build e2e

package api_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/yeniklas/gokapi-tui/internal/api"
	"github.com/yeniklas/gokapi-tui/internal/config"
	"github.com/yeniklas/gokapi-tui/internal/model"
)

func loadClient(t *testing.T) *api.Client {
	t.Helper()
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("cannot resolve config path: %v", err)
	}
	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Skipf("skipping e2e test: %v", err)
	}
	return api.New(cfg.ServerURL, cfg.APIKey)
}

func loadClientAndURL(t *testing.T) (*api.Client, string) {
	t.Helper()
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("cannot resolve config path: %v", err)
	}
	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Skipf("skipping e2e test: %v", err)
	}
	return api.New(cfg.ServerURL, cfg.APIKey), cfg.ServerURL
}

func uploadTempFile(t *testing.T, client *api.Client, params model.UploadParams) model.UploadResponse {
	t.Helper()
	f, err := os.CreateTemp("", "gokapi-tui-e2e-*.txt")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	f.WriteString("gokapi-tui e2e test file")
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	resp, err := client.UploadFile(context.Background(), f.Name(), params)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteFile(context.Background(), resp.FileInfo.Id)
	})
	return resp
}

func TestE2E_ListFiles(t *testing.T) {
	client := loadClient(t)
	files, err := client.ListFiles(context.Background())
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if files == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestE2E_UploadListDelete(t *testing.T) {
	client := loadClient(t)
	params := model.UploadParams{AllowedDownloads: 1, ExpiryDays: 1}
	resp := uploadTempFile(t, client, params)
	id := resp.FileInfo.Id
	if id == "" {
		t.Fatal("upload returned empty file ID")
	}

	files, err := client.ListFiles(context.Background())
	if err != nil {
		t.Fatalf("ListFiles after upload: %v", err)
	}
	found := false
	for _, f := range files {
		if f.Id == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("uploaded file %q not found in ListFiles", id)
	}

	if err := client.DeleteFile(context.Background(), id); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	files, err = client.ListFiles(context.Background())
	if err != nil {
		t.Fatalf("ListFiles after delete: %v", err)
	}
	for _, f := range files {
		if f.Id == id {
			t.Errorf("deleted file %q still appears in ListFiles", id)
		}
	}
}

func TestE2E_UploadParams(t *testing.T) {
	client := loadClient(t)
	params := model.UploadParams{AllowedDownloads: 2, ExpiryDays: 3}
	resp := uploadTempFile(t, client, params)

	if resp.FileInfo.UnlimitedDownloads {
		t.Error("expected UnlimitedDownloads=false for AllowedDownloads=2")
	}
	if resp.FileInfo.UnlimitedTime {
		t.Error("expected UnlimitedTime=false for ExpiryDays=3")
	}
	if resp.FileInfo.DownloadsRemaining != 2 {
		t.Errorf("DownloadsRemaining = %d, want 2", resp.FileInfo.DownloadsRemaining)
	}
}

func TestE2E_FileRequest_CreateListDelete(t *testing.T) {
	client := loadClient(t)
	params := model.CreateFileRequestParams{
		Name:     "e2e-test-request",
		MaxFiles: 1,
	}
	fr, err := client.CreateFileRequest(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateFileRequest: %v", err)
	}
	if fr.Id == "" {
		t.Fatal("CreateFileRequest returned empty ID")
	}
	t.Cleanup(func() {
		_ = client.DeleteFileRequest(context.Background(), fr.Id)
	})

	frs, err := client.ListFileRequests(context.Background())
	if err != nil {
		t.Fatalf("ListFileRequests: %v", err)
	}
	found := false
	for _, r := range frs {
		if r.Id == fr.Id {
			found = true
			if r.Name != "e2e-test-request" {
				t.Errorf("Name = %q, want %q", r.Name, "e2e-test-request")
			}
			break
		}
	}
	if !found {
		t.Errorf("created file request %q not found in ListFileRequests", fr.Id)
	}

	if err := client.DeleteFileRequest(context.Background(), fr.Id); err != nil {
		t.Fatalf("DeleteFileRequest: %v", err)
	}

	frs, err = client.ListFileRequests(context.Background())
	if err != nil {
		t.Fatalf("ListFileRequests after delete: %v", err)
	}
	for _, r := range frs {
		if r.Id == fr.Id {
			t.Errorf("deleted file request %q still appears in ListFileRequests", fr.Id)
		}
	}
}

func TestE2E_FileRequest_UploadURL(t *testing.T) {
	client, serverURL := loadClientAndURL(t)
	params := model.CreateFileRequestParams{Name: "e2e-url-test"}
	fr, err := client.CreateFileRequest(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateFileRequest: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteFileRequest(context.Background(), fr.Id)
	})

	if fr.Id == "" {
		t.Fatal("file request has empty Id")
	}
	if fr.ApiKey == "" {
		t.Fatal("file request has empty ApiKey — upload URL would be broken")
	}

	wantURL := serverURL + "/publicUpload?id=" + fr.Id + "&key=" + fr.ApiKey
	_ = wantURL // the TUI builds this; verify the components are non-empty
	t.Logf("upload URL: %s", wantURL)
}

func TestE2E_AuthFailure(t *testing.T) {
	_, serverURL := loadClientAndURL(t)
	badClient := api.New(serverURL, "invalid-api-key")
	_, err := badClient.ListFiles(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid API key, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401, got: %v", err)
	}
}
