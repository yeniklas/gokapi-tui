package api

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeniklas/gokapi-tui/internal/model"
)

func TestNew_TrimsTrailingSlash(t *testing.T) {
	c := New("http://example.com/", "key")
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
}

func TestListFiles_HappyPath(t *testing.T) {
	files := []model.GokapiFile{
		{Id: "abc", Name: "report.pdf", Size: "1.2 MB"},
		{Id: "def", Name: "photo.jpg", Size: "3.4 MB"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/files/list" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("apikey") != "testkey" {
			t.Errorf("missing or wrong apikey header: %q", r.Header.Get("apikey"))
		}
		json.NewEncoder(w).Encode(files)
	}))
	defer srv.Close()

	c := New(srv.URL, "testkey")
	got, err := c.ListFiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d files, want 2", len(got))
	}
	if got[0].Id != "abc" || got[1].Name != "photo.jpg" {
		t.Errorf("unexpected file data: %+v", got)
	}
}

func TestListFiles_NullResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("null"))
	}))
	defer srv.Close()

	c := New(srv.URL, "key")
	got, err := c.ListFiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice for null response, got %d items", len(got))
	}
}

func TestListFiles_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	c := New(srv.URL, "badkey")
	_, err := c.ListFiles(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestUploadFile_HappyPath(t *testing.T) {
	expected := model.UploadResponse{
		Result:   "OK",
		FileInfo: model.GokapiFile{Id: "xyz", Name: "test.txt", UrlDownload: "http://example.com/d/xyz"},
	}

	var gotDownloads, gotExpiry, gotPassword, gotFilename string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/files/add" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("apikey") != "testkey" {
			t.Errorf("missing apikey header")
		}
		r.ParseMultipartForm(1 << 20)
		gotDownloads = r.FormValue("allowedDownloads")
		gotExpiry = r.FormValue("expiryDays")
		gotPassword = r.FormValue("password")
		_, fh, _ := r.FormFile("file")
		if fh != nil {
			gotFilename = fh.Filename
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	// create temp file to upload
	f, err := os.CreateTemp("", "upload-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello world")
	f.Close()
	defer os.Remove(f.Name())

	c := New(srv.URL, "testkey")
	params := model.UploadParams{AllowedDownloads: 3, ExpiryDays: 14, Password: "secret"}
	resp, err := c.UploadFile(context.Background(), f.Name(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FileInfo.Id != "xyz" {
		t.Errorf("FileInfo.Id = %q, want %q", resp.FileInfo.Id, "xyz")
	}
	if gotDownloads != "3" {
		t.Errorf("allowedDownloads field = %q, want %q", gotDownloads, "3")
	}
	if gotExpiry != "14" {
		t.Errorf("expiryDays field = %q, want %q", gotExpiry, "14")
	}
	if gotPassword != "secret" {
		t.Errorf("password field = %q, want %q", gotPassword, "secret")
	}
	if gotFilename != filepath.Base(f.Name()) {
		t.Errorf("file field filename = %q, want %q", gotFilename, filepath.Base(f.Name()))
	}
}

func TestUploadFile_NonexistentFile(t *testing.T) {
	c := New("http://localhost", "key")
	_, err := c.UploadFile(context.Background(), "/nonexistent/file.txt", model.UploadParams{})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestUploadFile_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	f, err := os.CreateTemp("", "upload-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	c := New(srv.URL, "badkey")
	_, err = c.UploadFile(context.Background(), f.Name(), model.UploadParams{AllowedDownloads: 1, ExpiryDays: 1})
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestDeleteFile_HappyPath(t *testing.T) {
	var gotID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/files/delete" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("apikey") != "testkey" {
			t.Errorf("missing apikey header")
		}
		gotID = r.Header.Get("id")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "testkey")
	if err := c.DeleteFile(context.Background(), "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != "abc123" {
		t.Errorf("id header = %q, want %q", gotID, "abc123")
	}
}

func TestDeleteFile_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(srv.URL, "key")
	err := c.DeleteFile(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

// verify the multipart body is well-formed
func TestUploadFile_MultipartBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}
		_, params, _ := mime.ParseMediaType(ct)
		mr := multipart.NewReader(r.Body, params["boundary"])
		fields := map[string]string{}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("reading multipart: %v", err)
			}
			data, _ := io.ReadAll(p)
			fields[p.FormName()] = string(data)
		}
		if fields["allowedDownloads"] != "2" {
			t.Errorf("allowedDownloads = %q, want %q", fields["allowedDownloads"], "2")
		}
		if fields["expiryDays"] != "10" {
			t.Errorf("expiryDays = %q, want %q", fields["expiryDays"], "10")
		}
		json.NewEncoder(w).Encode(model.UploadResponse{})
	}))
	defer srv.Close()

	f, err := os.CreateTemp("", "*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("data")
	f.Close()
	defer os.Remove(f.Name())

	c := New(srv.URL, "key")
	c.UploadFile(context.Background(), f.Name(), model.UploadParams{AllowedDownloads: 2, ExpiryDays: 10})
}
