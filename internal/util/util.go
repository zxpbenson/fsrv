package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// HumanReadableSize converts bytes to human readable format
func HumanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// CheckAndCreateDir checks if a directory exists and creates it if it doesn't
func CheckAndCreateDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		fmt.Printf("Directory created: %s\n", dir)
	} else if err != nil {
		return fmt.Errorf("failed to check directory: %w", err)
	} else {
		fmt.Printf("Directory already exists: %s\n", dir)
	}
	return nil
}

// PrepareTmpDir prepares the temporary directory for file uploads
func PrepareTmpDir() (string, error) {
	executablePath, err := os.Executable() // create store dir base workdir
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Println("Current executable path:", executablePath)
	tmpDir := filepath.Join(filepath.Dir(executablePath), "tmp")
	fmt.Println("Current tmp dir path:", tmpDir)

	if err := CheckAndCreateDir(tmpDir); err != nil {
		return "", err
	}

	return tmpDir, nil
}

// SafeFileName returns a safe filename by extracting the base name
func SafeFileName(filename string) string {
	return filepath.Base(filename)
}
