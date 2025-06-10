package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	APP_NAME   = "bot-notify"
	DEPLOY_DIR = "deploy"
	OUTPUT_DIR = "deploy/output"
	VERSION    = "1.0.0"
)

func main() {
	fmt.Println("🚀 Mempersiapkan aplikasi untuk deployment...")
	startTime := time.Now()

	// 1. Membersihkan build artifacts sebelumnya
	cleanup()

	// 2. Build aplikasi dengan flag optimasi
	buildApp()

	// 3. Menyiapkan struktur direktori
	prepareDirs()

	// 4. Menyalin file-file yang diperlukan
	copyFiles()

	// 5. Kompresi ke dalam ZIP
	zipFiles()

	elapsedTime := time.Since(startTime)
	fmt.Printf("\n✅ Proses deployment selesai dalam %s\n", elapsedTime.Round(time.Millisecond))
	fmt.Printf("📦 File deployment tersedia di: %s/%s-%s.zip\n", DEPLOY_DIR, APP_NAME, VERSION)
}

// cleanup membersihkan build artifacts sebelumnya
func cleanup() {
	fmt.Println("\n🧹 Membersihkan build artifacts sebelumnya...")

	// Hapus binary aplikasi jika sudah ada
	if _, err := os.Stat(APP_NAME); err == nil {
		os.Remove(APP_NAME)
		fmt.Printf("   ✓ Binary %s dihapus\n", APP_NAME)
	}

	if _, err := os.Stat(APP_NAME + ".exe"); err == nil {
		os.Remove(APP_NAME + ".exe")
		fmt.Printf("   ✓ Binary %s.exe dihapus\n", APP_NAME)
	}

	// Hapus direktori tmp jika ada
	if _, err := os.Stat("tmp"); err == nil {
		os.RemoveAll("tmp")
		fmt.Println("   ✓ Direktori tmp dihapus")
	}

	// Hapus direktori output jika ada
	if _, err := os.Stat(OUTPUT_DIR); err == nil {
		os.RemoveAll(OUTPUT_DIR)
		fmt.Println("   ✓ Direktori output sebelumnya dihapus")
	}
}

// buildApp melakukan kompilasi aplikasi dengan flag optimasi - DIMODIFIKASI untuk tidak menggunakan exec.Command
func buildApp() {
	fmt.Println("\n🔨 Membangun aplikasi dengan optimasi...")

	// Deteksi sistem target (Linux untuk server)
	targetOS := "linux" // Default untuk server
	targetArch := "amd64"

	// Menentukan nama output berdasarkan target OS
	outputName := APP_NAME
	if targetOS == "windows" {
		outputName += ".exe"
	}

	fmt.Printf("   ℹ️ Target deployment: %s/%s\n", targetOS, targetArch)
	fmt.Printf("   ⚠️ Silakan jalankan perintah build manual berikut:\n")
	fmt.Printf("   GOOS=%s GOARCH=%s go build -ldflags=\"-s -w\" -o %s ./cmd/main.go\n\n", targetOS, targetArch, outputName)

	// Kita tidak menggunakan exec.Command lagi karena itu yang menyebabkan error
	fmt.Println("   🔄 Melewati proses build otomatis, silakan build manual...")
}

// prepareDirs menyiapkan struktur direktori untuk deployment
func prepareDirs() {
	fmt.Println("\n📂 Menyiapkan struktur direktori...")

	// Struktur direktori untuk deployment
	dirs := []string{
		OUTPUT_DIR,
		filepath.Join(OUTPUT_DIR, "config"),
		filepath.Join(OUTPUT_DIR, "internal", "web", "view"),
		filepath.Join(OUTPUT_DIR, "static"),
		filepath.Join(OUTPUT_DIR, "data", "whatsapp"),
		filepath.Join(OUTPUT_DIR, "data", "qrcodes"),
		filepath.Join(OUTPUT_DIR, "data", "sessions"),
		filepath.Join(OUTPUT_DIR, "data", "storage"),
		filepath.Join(OUTPUT_DIR, "logs"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("❌ Error saat membuat direktori %s: %v\n", dir, err)
			os.Exit(1)
		}
		fmt.Printf("   ✓ Direktori %s dibuat\n", dir)
	}
}

// copyFiles menyalin file-file yang diperlukan ke direktori deployment
func copyFiles() {
	fmt.Println("\n📋 Menyalin file-file aplikasi...")

	// Binary aplikasi
	binaryName := APP_NAME
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	if _, err := os.Stat(binaryName); err == nil {
		copyFile(binaryName, filepath.Join(OUTPUT_DIR, binaryName))
		fmt.Printf("   ✓ Binary %s disalin\n", binaryName)
	} else {
		fmt.Printf("⚠️ Binary %s tidak ditemukan\n", binaryName)
	}

	// Konfigurasi - khusus menyalin config-example.yaml dan menamainya menjadi config.yaml
	configExamplePath := filepath.Join("config", "config-example.yaml")
	configOutputPath := filepath.Join(OUTPUT_DIR, "config", "config.yaml")

	if _, err := os.Stat(configExamplePath); err == nil {
		copyFile(configExamplePath, configOutputPath)
		fmt.Println("   ✓ config-example.yaml disalin sebagai config.yaml")
	} else {
		fmt.Printf("⚠️ File %s tidak ditemukan\n", configExamplePath)
	}

	// View templates
	copyDir(filepath.Join("internal", "web", "view"), filepath.Join(OUTPUT_DIR, "internal", "web", "view"))

	// Static files
	copyDir("static", filepath.Join(OUTPUT_DIR, "static"))

	// README atau instruksi jika ada
	if _, err := os.Stat("README.md"); err == nil {
		copyFile("README.md", filepath.Join(OUTPUT_DIR, "README.md"))
		fmt.Println("   ✓ README.md disalin")
	}
}

// zipFiles mengompres direktori output ke dalam format ZIP
func zipFiles() {
	fmt.Println("\n🗜️ Mengompres file untuk deployment...")

	// Membuat direktori deploy jika belum ada
	if err := os.MkdirAll(DEPLOY_DIR, 0755); err != nil {
		fmt.Printf("❌ Error saat membuat direktori %s: %v\n", DEPLOY_DIR, err)
		os.Exit(1)
	}

	// Nama file ZIP
	zipFileName := fmt.Sprintf("%s/%s-%s.zip", DEPLOY_DIR, APP_NAME, VERSION)

	// Buat file ZIP
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		fmt.Printf("❌ Error saat membuat file ZIP: %v\n", err)
		os.Exit(1)
	}
	defer zipFile.Close()

	// Membuat writer ZIP
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Mengompres semua file dalam direktori output
	err = filepath.Walk(OUTPUT_DIR, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip direktori induk
		if path == OUTPUT_DIR {
			return nil
		}

		// Buat header ZIP
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set nama relatif dalam ZIP
		relPath, err := filepath.Rel(OUTPUT_DIR, path)
		if err != nil {
			return err
		}

		// Gunakan forward slash untuk path di ZIP, termasuk di Windows
		header.Name = strings.ReplaceAll(relPath, "\\", "/")

		// Set metode kompresi
		header.Method = zip.Deflate

		// Untuk direktori, tambahkan trailing slash
		if info.IsDir() {
			header.Name += "/"
			// Buat entry dalam ZIP
			_, err = zipWriter.CreateHeader(header)
			return err
		}

		// Buat entry dalam ZIP untuk file
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Untuk file, salin konten
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		fmt.Printf("   ✓ File %s ditambahkan ke ZIP\n", relPath)
		return err
	})

	if err != nil {
		fmt.Printf("❌ Error saat membuat ZIP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   ✓ File ZIP dibuat: %s\n", zipFileName)
}

// copyFile menyalin file tunggal dari source ke destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// copyDir menyalin direktori dari source ke destination
func copyDir(src, dst string) error {
	// Periksa apakah direktori sumber ada
	srcInfo, err := os.Stat(src)
	if err != nil {
		fmt.Printf("⚠️ Direktori/file %s tidak ditemukan: %v\n", src, err)
		return nil // Biarkan proses tetap berlanjut
	}

	// Jika source bukan direktori, salin sebagai file
	if !srcInfo.IsDir() {
		return copyFile(src, dst)
	}

	// Buat direktori tujuan jika belum ada
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Baca isi direktori
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Salin setiap entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Rekursif untuk subdirektori
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Salin file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}

		fmt.Printf("   ✓ %s disalin\n", srcPath)
	}

	return nil
}
