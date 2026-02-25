package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHumanReadableSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "bytes",
			size: 100,
			want: "100 B",
		},
		{
			name: "kilobytes",
			size: 1024,
			want: "1.0 KB",
		},
		{
			name: "megabytes",
			size: 1024 * 1024,
			want: "1.0 MB",
		},
		{
			name: "gigabytes",
			size: 1024 * 1024 * 1024,
			want: "1.0 GB",
		},
		{
			name: "terabytes",
			size: 1024 * 1024 * 1024 * 1024,
			want: "1.0 TB",
		},
		{
			name: "petabytes",
			size: 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1.0 PB",
		},
		{
			name: "exabytes",
			size: 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1.0 EB",
		},
		{
			name: "zero",
			size: 0,
			want: "0 B",
		},
		{
			name: "fractional KB",
			size: 1536,
			want: "1.5 KB",
		},
		{
			name: "fractional MB",
			size: 1572864,
			want: "1.5 MB",
		},
		{
			name: "large size",
			size: 5 * 1024 * 1024 * 1024,
			want: "5.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanReadableSize(tt.size); got != tt.want {
				t.Errorf("HumanReadableSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAndCreateDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := filepath.Join(os.TempDir(), "fsrv-test")
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{
			name:    "create new directory",
			dir:     tmpDir,
			wantErr: false,
		},
		{
			name:    "directory already exists",
			dir:     tmpDir,
			wantErr: false,
		},
		{
			name:    "create nested directory",
			dir:     filepath.Join(tmpDir, "nested", "dir"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAndCreateDir(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAndCreateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify directory exists
			if !tt.wantErr {
				info, err := os.Stat(tt.dir)
				if err != nil {
					t.Errorf("Directory does not exist: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("Path is not a directory")
				}
			}
		})
	}
}

func TestSafeFileName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "test.txt",
			want:     "test.txt",
		},
		{
			name:     "path with directory",
			filename: "/path/to/file.txt",
			want:     "file.txt",
		},
		{
			name:     "relative path",
			filename: "../file.txt",
			want:     "file.txt",
		},
		{
			name:     "complex path",
			filename: "/a/b/c/d/e/f/g/file.txt",
			want:     "file.txt",
		},
		{
			name:     "filename with dots",
			filename: "file.name.with.dots.txt",
			want:     "file.name.with.dots.txt",
		},
		{
			name:     "just filename",
			filename: "file",
			want:     "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeFileName(tt.filename); got != tt.want {
				t.Errorf("SafeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrepareTmpDir(t *testing.T) {
	// This test is tricky because PrepareTmpDir uses os.Executable
	// We'll just test that it returns a valid path
	tmpDir, err := PrepareTmpDir()
	if err != nil {
		t.Errorf("PrepareTmpDir() error = %v", err)
		return
	}

	if tmpDir == "" {
		t.Error("PrepareTmpDir() returned empty string")
	}

	// Verify the directory exists
	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Errorf("Temporary directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("Path is not a directory")
	}
}

// BenchmarkHumanReadableSize benchmarks the HumanReadableSize function
func BenchmarkHumanReadableSize(b *testing.B) {
	sizes := []int64{
		100,
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		1024 * 1024 * 1024 * 1024,
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				HumanReadableSize(size)
			}
		})
	}
}

// BenchmarkSafeFileName benchmarks the SafeFileName function
func BenchmarkSafeFileName(b *testing.B) {
	filenames := []string{
		"test.txt",
		"/path/to/file.txt",
		"../file.txt",
		"/a/b/c/d/e/f/g/file.txt",
	}

	for _, filename := range filenames {
		b.Run(filename, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				SafeFileName(filename)
			}
		})
	}
}

// ExampleHumanReadableSize demonstrates how to convert bytes to human readable format
func ExampleHumanReadableSize() {
	size := int64(1024 * 1024 * 512) // 512 MB
	fmt.Println(HumanReadableSize(size))
	// Output: 512.0 MB
}

// ExampleSafeFileName demonstrates how to safely extract filename from a path
func ExampleSafeFileName() {
	path := "/path/to/important/file.txt"
	fmt.Println(SafeFileName(path))
	// Output: file.txt
}
