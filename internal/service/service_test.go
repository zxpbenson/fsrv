package service

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fsrv/internal/config"
)

func setupTestService(t *testing.T) (*Service, string) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "fsrv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    tmpDir,
		Max:      32,
	}

	svc := New(cfg)
	return svc, tmpDir
}

func cleanupTestService(t *testing.T, tmpDir string) {
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Logf("Failed to cleanup temp dir: %v", err)
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	if svc == nil {
		t.Error("New() returned nil")
	}

	if svc.cfg != cfg {
		t.Error("New() did not set config correctly")
	}
}

func TestService_ListFiles(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		// Add delay to ensure different modification times
		time.Sleep(10 * time.Millisecond)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	files, err := svc.ListFiles()
	if err != nil {
		t.Errorf("ListFiles() error = %v", err)
		return
	}

	if len(files) != len(testFiles) {
		t.Errorf("ListFiles() returned %d files, want %d", len(files), len(testFiles))
	}

	// Verify files are sorted by modification time (newest first)
	for i := 0; i < len(files)-1; i++ {
		if files[i].Filename == testFiles[0] {
			// First file should be file1.txt (oldest)
			t.Error("Files should be sorted by modification time (newest first)")
		}
	}
}

func TestService_ListFiles_Empty(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	files, err := svc.ListFiles()
	if err != nil {
		t.Errorf("ListFiles() error = %v", err)
		return
	}

	if len(files) != 0 {
		t.Errorf("ListFiles() returned %d files, want 0", len(files))
	}
}

func TestService_UploadFile(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	filename := "test.txt"
	content := []byte("test content")
	reader := bytes.NewReader(content)

	size, err := svc.UploadFile(filename, reader)
	if err != nil {
		t.Errorf("UploadFile() error = %v", err)
		return
	}

	if size != int64(len(content)) {
		t.Errorf("UploadFile() returned size %d, want %d", size, len(content))
	}

	// Verify file exists
	path := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Uploaded file does not exist")
	}

	// Verify file content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read uploaded file: %v", err)
	}

	if !bytes.Equal(data, content) {
		t.Error("Uploaded file content does not match")
	}
}

func TestService_UploadFile_AlreadyExists(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	filename := "test.txt"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	reader := bytes.NewReader([]byte("new content"))
	_, err := svc.UploadFile(filename, reader)
	if err == nil {
		t.Error("UploadFile() should return error when file already exists")
	}
}

func TestService_UploadFile_PathTraversal(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	// Try to upload with path traversal
	filename := "../test.txt"
	content := []byte("test content")
	reader := bytes.NewReader(content)

	_, err := svc.UploadFile(filename, reader)
	if err != nil {
		t.Errorf("UploadFile() error = %v", err)
	}

	// Verify file was saved in the correct directory (not parent)
	path := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("File was not saved in the correct directory")
	}
}

func TestService_DeleteFile(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	filename := "test.txt"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := svc.DeleteFile(filename)
	if err != nil {
		t.Errorf("DeleteFile() error = %v", err)
		return
	}

	// Verify file was deleted
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestService_DeleteFile_NotExists(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	err := svc.DeleteFile("nonexistent.txt")
	if err == nil {
		t.Error("DeleteFile() should return error when file does not exist")
	}
}

func TestService_DeleteFile_Directory(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	// Create a directory
	dirname := "testdir"
	path := filepath.Join(tmpDir, dirname)
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err := svc.DeleteFile(dirname)
	if err == nil {
		t.Error("DeleteFile() should return error when trying to delete a directory")
	}
}

func TestService_GetFilePath(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	filename := "test.txt"
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := svc.GetFilePath(filename)
	if err != nil {
		t.Errorf("GetFilePath() error = %v", err)
		return
	}

	if result != path {
		t.Errorf("GetFilePath() returned %s, want %s", result, path)
	}
}

func TestService_GetFilePath_NotExists(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	_, err := svc.GetFilePath("nonexistent.txt")
	if err == nil {
		t.Error("GetFilePath() should return error when file does not exist")
	}
}

func TestService_GetFilePath_Directory(t *testing.T) {
	svc, tmpDir := setupTestService(t)
	defer cleanupTestService(t, tmpDir)

	dirname := "testdir"
	path := filepath.Join(tmpDir, dirname)
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	_, err := svc.GetFilePath(dirname)
	if err == nil {
		t.Error("GetFilePath() should return error when path is a directory")
	}
}

func TestService_GetMaxUploadSize(t *testing.T) {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	expected := int64(1 << 32) // 4GB

	if size := svc.GetMaxUploadSize(); size != expected {
		t.Errorf("GetMaxUploadSize() = %d, want %d", size, expected)
	}
}

func TestService_GetMaxUploadSizeHuman(t *testing.T) {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	expected := "4.0 GB"

	if size := svc.GetMaxUploadSizeHuman(); size != expected {
		t.Errorf("GetMaxUploadSizeHuman() = %s, want %s", size, expected)
	}
}

func TestService_IsDeleteEnabled(t *testing.T) {
	tests := []struct {
		name     string
		delAble  bool
		expected bool
	}{
		{
			name:     "delete enabled",
			delAble:  true,
			expected: true,
		},
		{
			name:     "delete disabled",
			delAble:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Port:     "8080",
				DelAble:  tt.delAble,
				Hostname: "localhost",
				Store:    "./store",
				Max:      32,
			}

			svc := New(cfg)
			if enabled := svc.IsDeleteEnabled(); enabled != tt.expected {
				t.Errorf("IsDeleteEnabled() = %v, want %v", enabled, tt.expected)
			}
		})
	}
}

func TestService_GetServerInfo(t *testing.T) {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "test-server",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	hostname, port, maxSize := svc.GetServerInfo()

	if hostname != "test-server" {
		t.Errorf("GetServerInfo() hostname = %s, want test-server", hostname)
	}

	if port != "8080" {
		t.Errorf("GetServerInfo() port = %s, want 8080", port)
	}

	if maxSize != "4.0 GB" {
		t.Errorf("GetServerInfo() maxSize = %s, want 4.0 GB", maxSize)
	}
}

func TestGetCurrentTime(t *testing.T) {
	timeStr := GetCurrentTime()
	if timeStr == "" {
		t.Error("GetCurrentTime() returned empty string")
	}

	// Try to parse the time to verify format
	_, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		t.Errorf("GetCurrentTime() returned invalid time format: %v", err)
	}
}

// BenchmarkService_ListFiles benchmarks the ListFiles method
func BenchmarkService_ListFiles(b *testing.B) {
	svc, tmpDir := setupTestService(b)
	defer cleanupTestService(b, tmpDir)

	// Create test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune(i))+".txt")
		if err := os.WriteFile(filename, []byte("test content"), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ListFiles()
	}
}

// BenchmarkService_UploadFile benchmarks the UploadFile method
func BenchmarkService_UploadFile(b *testing.B) {
	svc, tmpDir := setupTestService(b)
	defer cleanupTestService(b, tmpDir)

	content := bytes.Repeat([]byte("test"), 1024*1024) // 4MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "test.txt"
		reader := bytes.NewReader(content)
		svc.UploadFile(filename, reader)
		os.Remove(filepath.Join(tmpDir, filename))
	}
}

// ExampleService_ListFiles demonstrates how to list files
func ExampleService_ListFiles() {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	files, err := svc.ListFiles()
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Printf("%s (%s)\n", file.Filename, file.Size)
	}
}

// ExampleService_UploadFile demonstrates how to upload a file
func ExampleService_UploadFile() {
	cfg := &config.Config{
		Port:     "8080",
		DelAble:  true,
		Hostname: "localhost",
		Store:    "./store",
		Max:      32,
	}

	svc := New(cfg)
	content := []byte("Hello, World!")
	reader := bytes.NewReader(content)

	size, err := svc.UploadFile("hello.txt", reader)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Uploaded %d bytes\n", size)
	// Output: Uploaded 13 bytes
}
