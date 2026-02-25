package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fsrv/internal/config"
	"fsrv/internal/service"
)

func setupTestHandler(t *testing.T) (*Handler, string) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "fsrv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create temporary templates directory
	templateDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create test templates
	templates := map[string]string{
		"files.html":  `{{.Title}}`,
		"info.html":   `{{.Title}}`,
		"upload.html": `{{.Title}}`,
	}

	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create template %s: %v", name, err)
		}
	}

	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    tmpDir,
		Max:      32,
	}

	svc := service.New(cfg)
	h, err := New(svc, os.DirFS(templateDir))
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	return h, tmpDir
}

func cleanupTestHandler(t *testing.T, tmpDir string) {
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Logf("Failed to cleanup temp dir: %v", err)
	}
}

func TestNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fsrv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create templates directory
	templateDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create test template
	if err := os.WriteFile(filepath.Join(templateDir, "files.html"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    tmpDir,
		Max:      32,
	}

	svc := service.New(cfg)
	h, err := New(svc, os.DirFS(templateDir))
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if h == nil {
		t.Error("New() returned nil")
	}

	if h.svc != svc {
		t.Error("New() did not set service correctly")
	}

	if h.templates == nil {
		t.Error("New() did not set templates")
	}
}

func TestHandler_UploadPage(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("GET", "/toUpload", nil)
	w := httptest.NewRecorder()

	h.UploadPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UploadPage() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "FSrv Upload") {
		t.Error("UploadPage() response does not contain expected title")
	}
}

func TestHandler_UploadPage_WrongMethod(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("POST", "/toUpload", nil)
	w := httptest.NewRecorder()

	h.UploadPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UploadPage() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "HTTP Method should be 'GET'") {
		t.Error("UploadPage() response does not contain error message")
	}
}

func TestHandler_UploadFile(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.UploadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UploadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "Uploaded file successfully") {
		t.Error("UploadFile() response does not contain success message")
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Uploaded file does not exist")
	}
}

func TestHandler_UploadFile_NoFile(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()

	h.UploadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UploadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No file selected") {
		t.Error("UploadFile() response does not contain error message")
	}
}

func TestHandler_UploadFile_WrongMethod(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("GET", "/upload", nil)
	w := httptest.NewRecorder()

	h.UploadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UploadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "HTTP Method should be 'POST'") {
		t.Error("UploadFile() response does not contain error message")
	}
}

func TestHandler_ListFiles(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt"}
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/files", nil)
	w := httptest.NewRecorder()

	h.ListFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListFiles() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "FSrv Files") {
		t.Error("ListFiles() response does not contain expected title")
	}
}

func TestHandler_ListFiles_Empty(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("GET", "/files", nil)
	w := httptest.NewRecorder()

	h.ListFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListFiles() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "FSrv Files") {
		t.Error("ListFiles() response does not contain expected title")
	}
}

func TestHandler_ListFiles_WrongMethod(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("POST", "/files", nil)
	w := httptest.NewRecorder()

	h.ListFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListFiles() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "HTTP Method should be 'GET'") {
		t.Error("ListFiles() response does not contain error message")
	}
}

func TestHandler_DeleteFile(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	// Create test file
	filename := "test.txt"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := httptest.NewRequest("GET", "/del?file="+filename, nil)
	w := httptest.NewRecorder()

	h.DeleteFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DeleteFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Deleted file successfully") {
		t.Error("DeleteFile() response does not contain success message")
	}

	// Verify file was deleted
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestHandler_DeleteFile_NotExists(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("GET", "/del?file=nonexistent.txt", nil)
	w := httptest.NewRecorder()

	h.DeleteFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DeleteFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "file does not exist") {
		t.Error("DeleteFile() response does not contain error message")
	}
}

func TestHandler_DeleteFile_WrongMethod(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("POST", "/del?file=test.txt", nil)
	w := httptest.NewRecorder()

	h.DeleteFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DeleteFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "HTTP Method should be 'GET'") {
		t.Error("DeleteFile() response does not contain error message")
	}
}

func TestHandler_DownloadFile(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	// Create test file
	filename := "test.txt"
	path := filepath.Join(tmpDir, filename)
	content := []byte("test content")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := httptest.NewRequest("GET", "/download?file="+filename, nil)
	w := httptest.NewRecorder()

	h.DownloadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DownloadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Check Content-Disposition header
	contentDisposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "attachment") {
		t.Error("DownloadFile() response does not contain Content-Disposition header")
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Errorf("DownloadFile() Content-Type = %s, want application/octet-stream", contentType)
	}

	// Check file content
	body := w.Body.Bytes()
	if !bytes.Equal(body, content) {
		t.Error("DownloadFile() response content does not match file content")
	}
}

func TestHandler_DownloadFile_NotExists(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("GET", "/download?file=nonexistent.txt", nil)
	w := httptest.NewRecorder()

	h.DownloadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DownloadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "file does not exist") {
		t.Error("DownloadFile() response does not contain error message")
	}
}

func TestHandler_DownloadFile_WrongMethod(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	req := httptest.NewRequest("POST", "/download?file=test.txt", nil)
	w := httptest.NewRecorder()

	h.DownloadFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DownloadFile() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "HTTP Method should be 'GET'") {
		t.Error("DownloadFile() response does not contain error message")
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, tmpDir := setupTestHandler(t)
	defer cleanupTestHandler(t, tmpDir)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Test that routes are registered
	routes := []string{
		"/toUpload",
		"/upload",
		"/files",
		"/download",
		"/del",
		"/",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// All routes should return a response (even if it's an error page)
		if w.Code == 0 {
			t.Errorf("Route %s is not registered", route)
		}
	}
}

// BenchmarkHandler_UploadFile benchmarks the UploadFile handler
func BenchmarkHandler_UploadFile(b *testing.B) {
	h, tmpDir := setupTestHandler(b)
	defer cleanupTestHandler(b, tmpDir)

	content := bytes.Repeat([]byte("test"), 1024*1024) // 4MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write(content)
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		h.UploadFile(w, req)

		// Clean up uploaded file
		os.Remove(filepath.Join(tmpDir, "test.txt"))
	}
}

// BenchmarkHandler_ListFiles benchmarks the ListFiles handler
func BenchmarkHandler_ListFiles(b *testing.B) {
	h, tmpDir := setupTestHandler(b)
	defer cleanupTestHandler(b, tmpDir)

	// Create test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune(i))+".txt")
		os.WriteFile(filename, []byte("test content"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/files", nil)
		w := httptest.NewRecorder()
		h.ListFiles(w, req)
	}
}

// ExampleHandler_UploadFile demonstrates how to upload a file via HTTP
func ExampleHandler_UploadFile() {
	// This is a simplified example showing the concept
	// In real usage, you would use an HTTP client to make the request

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("Hello, World!"))
	writer.Close()

	// Create request
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// In real usage, you would send this request to the server
	_ = req
}
