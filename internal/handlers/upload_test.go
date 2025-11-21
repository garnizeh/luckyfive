package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/garnizeh/luckyfive/internal/services"
)

func TestUploadResults_Success(t *testing.T) {
	logger := createTestLogger()

	uploadSvc := services.NewUploadService(logger)
	tmp := t.TempDir()
	uploadSvc.SetTempDir(tmp)

	// Create multipart form with a small dummy .xlsx file
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "test.xlsx")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	// Write some bytes (UploadService doesn't validate XLSX contents)
	if _, err := io.Copy(fw, bytes.NewReader([]byte("dummy xlsx content"))); err != nil {
		t.Fatalf("writing file part failed: %v", err)
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := UploadResults(uploadSvc, logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		ArtifactID string `json:"artifact_id"`
		Filename   string `json:"filename"`
		Size       int64  `json:"size"`
		Message    string `json:"message"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Filename != "test.xlsx" {
		t.Fatalf("expected filename test.xlsx, got %s", resp.Filename)
	}
	if resp.ArtifactID == "" {
		t.Fatalf("expected artifact id to be set")
	}

	// Verify file exists on disk
	files, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	found := false
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".xlsx" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected uploaded file in temp dir")
	}
}

func TestUploadResults_MethodNotAllowed(t *testing.T) {
	logger := createTestLogger()
	uploadSvc := services.NewUploadService(logger)

	handler := UploadResults(uploadSvc, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/results/upload", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
