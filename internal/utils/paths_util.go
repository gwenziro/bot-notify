package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// ProjectRoot adalah path absolut ke direktori root proyek
	ProjectRoot string
)

func init() {
	// Deteksi project root secara otomatis
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Tidak dapat mendeteksi path proyek")
	}

	// Path ke direktori utils/
	dir := filepath.Dir(filename)

	// paths.go -> utils/ -> internal/ -> root/
	ProjectRoot = filepath.Clean(filepath.Join(dir, "..", ".."))

	// Override dengan environment variable jika ada
	if envRoot := os.Getenv("APP_ROOT"); envRoot != "" {
		ProjectRoot = envRoot
	}
}

// GetProjectDir mengembalikan path absolut ke direktori proyek
func GetProjectDir() string {
	return ProjectRoot
}

// EnsureDirectoryExists memastikan direktori ada, jika tidak maka akan dibuat
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// IsSubdirectory memeriksa apakah path adalah subdirektori dari parent
func IsSubdirectory(parent, path string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && rel != "."
}

// ResolvePath mengonversi path relatif menjadi absolut
func ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(ProjectRoot, path)
}

// EnsureProjectStructure memastikan struktur direktori proyek sudah benar
func EnsureProjectStructure() error {
	dirs := []string{
		filepath.Join(ProjectRoot, "data"),
		filepath.Join(ProjectRoot, "config"),
		filepath.Join(ProjectRoot, "logs"),
		filepath.Join(ProjectRoot, "tmp"),
	}

	for _, dir := range dirs {
		if err := EnsureDirectoryExists(dir); err != nil {
			return err
		}
	}

	return nil
}

// FileExists memeriksa apakah file ada
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists memeriksa apakah direktori ada
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
