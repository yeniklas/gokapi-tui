package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yeniklas/gokapi-tui/internal/model"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return resp, nil
}

func (c *Client) ListFiles(ctx context.Context) ([]model.GokapiFile, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/files/list", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var files []model.GokapiFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decoding file list: %w", err)
	}
	if files == nil {
		return []model.GokapiFile{}, nil
	}
	return files, nil
}

func (c *Client) UploadFile(ctx context.Context, path string, params model.UploadParams) (model.UploadResponse, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.UploadResponse{}, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	_ = mw.WriteField("allowedDownloads", strconv.Itoa(params.AllowedDownloads))
	_ = mw.WriteField("expiryDays", strconv.Itoa(params.ExpiryDays))
	_ = mw.WriteField("password", params.Password)

	fw, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return model.UploadResponse{}, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return model.UploadResponse{}, err
	}
	mw.Close()

	resp, err := c.do(ctx, http.MethodPost, "/api/files/add", &buf, mw.FormDataContentType())
	if err != nil {
		return model.UploadResponse{}, err
	}
	defer resp.Body.Close()

	var result model.UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.UploadResponse{}, fmt.Errorf("decoding upload response: %w", err)
	}
	return result, nil
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/files/delete", nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("id", id)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}
