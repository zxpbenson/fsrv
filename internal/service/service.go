package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"fsrv/internal/config"
	"fsrv/internal/util"
)

// File represents a file in the store
type File struct {
	Filename     string
	DownloadLink string
	Size         string
	ModifyTime   string
	Curl         string
}

// Service handles file operations
type Service struct {
	cfg *config.Config
	mu  sync.RWMutex
}

// New creates a new file service
func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// ListFiles returns a list of all files in the store directory
func (s *Service) ListFiles() ([]File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir, err := os.Open(s.cfg.Store)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	var result []File
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			downloadURL := fmt.Sprintf("%s/download?file=%s", s.getURLRoot(), fileName)

			result = append(result, File{
				Filename:     fileName,
				DownloadLink: downloadURL,
				Size:         util.HumanReadableSize(file.Size()),
				ModifyTime:   file.ModTime().Format("2006-01-02 15:04:05"),
				Curl:         fmt.Sprintf("curl -L -o '%s' '%s'", fileName, downloadURL),
			})
		}
	}

	return result, nil
}

// UploadFile saves an uploaded file to the store directory
func (s *Service) UploadFile(filename string, src io.Reader) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	safeFilename := util.SafeFileName(filename)
	fullPath := filepath.Join(s.cfg.Store, safeFilename)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return 0, fmt.Errorf("file already exists: '%s'", safeFilename)
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file with buffer
	buffer := make([]byte, 1024*1024) // 1MB buffer
	size, err := io.CopyBuffer(dst, src, buffer)
	if err != nil {
		return 0, fmt.Errorf("failed to save file: %w", err)
	}

	return size, nil
}

// DeleteFile removes a file from the store directory
func (s *Service) DeleteFile(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	safeFilename := util.SafeFileName(filename)
	filePath := filepath.Join(s.cfg.Store, safeFilename)

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: '%s'", safeFilename)
	}
	if err != nil {
		return fmt.Errorf("failed to check file: %w", err)
	}

	// Cannot delete directories
	if info.IsDir() {
		return fmt.Errorf("cannot delete directory: '%s'", safeFilename)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// OpenFile safely opens a file for reading under a read lock.
//
// Concurrency Safety Note:
// This method uses a read lock (RLock) to protect the file opening process.
// Once the file is successfully opened and the handle (*os.File) is returned,
// the lock is released. This is safe because:
// 1. On Unix-like systems, an open file handle remains valid even if the underlying
//    directory entry is removed (e.g., via DeleteFile). The data can still be read fully.
// 2. On Windows, the OS prevents deleting a file that is currently open. DeleteFile
//    will fail, effectively protecting the ongoing download.
// This design minimizes lock contention by only holding the lock during the Open operation,
// avoiding blocking other operations (like Upload/Delete) during long downloads.
func (s *Service) OpenFile(filename string) (*os.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	safeFilename := util.SafeFileName(filename)
	filePath := filepath.Join(s.cfg.Store, safeFilename)

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: '%s'", safeFilename)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check file: %w", err)
	}

	// Cannot download directories
	if info.IsDir() {
		return nil, fmt.Errorf("cannot download directory: '%s'", safeFilename)
	}

	// Open the file while holding the read lock
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetMaxUploadSize returns the maximum upload size in bytes
func (s *Service) GetMaxUploadSize() int64 {
	return int64(1) << s.cfg.Max
}

// GetMaxUploadSizeHuman returns the maximum upload size in human readable format
func (s *Service) GetMaxUploadSizeHuman() string {
	return util.HumanReadableSize(int64(1) << s.cfg.Max)
}

// IsDeleteEnabled returns whether delete functionality is enabled
func (s *Service) IsDeleteEnabled() bool {
	return s.cfg.DelAble
}

// getURLRoot returns the base URL for the server
func (s *Service) getURLRoot() string {
	return fmt.Sprintf("http://%s:%s", s.cfg.Hostname, s.cfg.Port)
}

// GetServerInfo returns server information for the upload page
func (s *Service) GetServerInfo() (hostname, port, maxSize string) {
	return s.cfg.Hostname, s.cfg.Port, s.GetMaxUploadSizeHuman()
}

// GetCurrentTime returns the current time in a formatted string
func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
