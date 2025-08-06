package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// ProjectRoot adalah path absolut ke direktori root proyek
	ProjectRoot string
	// AppMode menyimpan mode aplikasi (development/production)
	AppMode string
)

func init() {
	// Panggil InitProjectRoot untuk menginisialisasi ProjectRoot dengan benar
	InitProjectRoot()

	// Log direktori yang terdeteksi untuk debugging
	fmt.Printf("Project root terdeteksi: %s\n", ProjectRoot)
	fmt.Printf("App mode: %s\n", AppMode)

	// Pastikan struktur direktori utama ada
	err := EnsureProjectStructure()
	if err != nil {
		fmt.Printf("PERINGATAN: Gagal membuat struktur direktori: %v\n", err)
	}
}

// InitProjectRoot menginisialisasi ProjectRoot dengan cara yang lebih cerdas
func InitProjectRoot() {
	// Deteksi mode aplikasi
	AppMode = os.Getenv("APP_MODE")
	if AppMode == "" {
		AppMode = "development" // Default ke development
	}

	// PRIORITAS 1: Cek apakah environment variable diset
	envRoot := os.Getenv("APP_ROOT")
	if envRoot != "" {
		ProjectRoot = envRoot
		fmt.Printf("Menggunakan APP_ROOT dari environment: %s\n", ProjectRoot)
		return
	}

	// PRIORITAS 2: Gunakan current working directory sebagai fallback
	cwd, err := os.Getwd()
	if err == nil {
		ProjectRoot = cwd
		fmt.Printf("Menggunakan current working directory: %s\n", ProjectRoot)
		return
	}

	// PRIORITAS 3: Deteksi berdasarkan runtime caller
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// Path ke direktori utils/
		dir := filepath.Dir(filename)
		// utils/ -> internal/ -> root/
		ProjectRoot = filepath.Clean(filepath.Join(dir, "..", ".."))
		fmt.Printf("Menggunakan path berdasarkan runtime.Caller: %s\n", ProjectRoot)
		return
	}

	// PRIORITAS 4: Deteksi berdasarkan lokasi binary
	execPath, err := os.Executable()
	if err == nil {
		ProjectRoot = filepath.Dir(execPath)
		fmt.Printf("Menggunakan lokasi binary: %s\n", ProjectRoot)
		return
	}

	// Fallback terakhir jika semua metode gagal
	ProjectRoot = "."
	fmt.Printf("Menggunakan fallback directory (.): %s\n", ProjectRoot)
}

// GetProjectDir mengembalikan path absolut ke direktori proyek
func GetProjectDir() string {
	return ProjectRoot
}

// EnsureDirectoryExists memastikan direktori ada, jika tidak maka akan dibuat
func EnsureDirectoryExists(path string) error {
	// Konversi path ke path absolut jika relatif
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(ProjectRoot, path)
	}

	// Log untuk debug
	fmt.Printf("Memastikan direktori ada: %s\n", absPath)

	err := os.MkdirAll(absPath, 0755)
	if err != nil {
		fmt.Printf("Error saat membuat direktori: %v\n", err)
	}
	return err
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

	resolved := filepath.Join(ProjectRoot, path)
	fmt.Printf("Path %s diresolve menjadi %s\n", path, resolved)
	return resolved
}

// EnsureProjectStructure memastikan struktur direktori proyek sudah benar
func EnsureProjectStructure() error {
	// Direktori utama yang diperlukan
	dirs := []string{
		filepath.Join(ProjectRoot, "data"),
		filepath.Join(ProjectRoot, "config"),
		filepath.Join(ProjectRoot, "logs"),
		filepath.Join(ProjectRoot, "tmp"),
		filepath.Join(ProjectRoot, "static"),
	}

	// Tambahkan direktori untuk file HTML - PENTING
	dirs = append(dirs,
		filepath.Join(ProjectRoot, "internal"),
		filepath.Join(ProjectRoot, "internal", "web"),
		filepath.Join(ProjectRoot, "internal", "web", "view"),
	)

	for _, dir := range dirs {
		fmt.Printf("Memeriksa direktori: %s\n", dir)
		if err := EnsureDirectoryExists(dir); err != nil {
			return err
		}
	}

	// Pastikan file HTML dasar ada - tambahkan pengecekan untuk file-file penting
	htmlFiles := []string{
		filepath.Join(ProjectRoot, "internal", "web", "view", "index.html"),
		filepath.Join(ProjectRoot, "internal", "web", "view", "dashboard.html"),
		filepath.Join(ProjectRoot, "internal", "web", "view", "login.html"),
		filepath.Join(ProjectRoot, "internal", "web", "view", "error.html"),
	}

	for _, file := range htmlFiles {
		if !FileExists(file) {
			fmt.Printf("PERINGATAN: File HTML penting tidak ditemukan: %s\n", file)
		} else {
			fmt.Printf("File HTML ditemukan: %s\n", file)
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

// RelativePath mengembalikan jalur relatif terhadap ProjectRoot
// atau mengubah path absolut Windows menjadi path relatif
func RelativePath(path string) string {
	// Jika path kosong, kembalikan ProjectRoot
	if path == "" {
		return ProjectRoot
	}

	// Jika path adalah absolute path di Windows (misalnya D:/...),
	// ambil bagian setelah drive letter atau root path
	if runtime.GOOS == "windows" && strings.Contains(path, ":") {
		parts := strings.Split(path, ":")
		if len(parts) > 1 {
			path = parts[1]
		}
	} else if strings.HasPrefix(path, "/") {
		// Untuk path absolut di Linux, hapus slash awal
		path = strings.TrimPrefix(path, "/")
	}

	// Hapus jalur proyek dari awal path jika ada
	path = strings.TrimPrefix(path, ProjectRoot)

	// Hapus slash awal jika ada
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "\\")

	// Buat jalur relatif
	return filepath.Join(ProjectRoot, path)
}
