package session

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/skip2/go-qrcode"
)

// QRHandler menangani operasi terkait QR code
type QRHandler struct {
	config     *config.Config
	logger     utils.LogrusEntry
	qrCodePath string
	timestamp  time.Time
}

// NewQRHandler membuat instance baru QRHandler
func NewQRHandler(cfg *config.Config, logger utils.LogrusEntry, qrCodePath string) *QRHandler {
	return &QRHandler{
		config:     cfg,
		logger:     logger.WithField("component", "qr-handler"),
		qrCodePath: qrCodePath,
	}
}

// SaveQRCode menyimpan QR code dari string ke file gambar
func (h *QRHandler) SaveQRCode(qrCode string) error {
	// Pastikan direktori QR code ada
	qrDir := filepath.Dir(h.qrCodePath)
	if err := os.MkdirAll(qrDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori QR code: %w", err)
	}

	// Bersihkan QR code string jika perlu
	qrCode = strings.TrimSpace(qrCode)

	// Buat QR code gambar
	qr, err := qrcode.New(qrCode, qrcode.Medium)
	if err != nil {
		return fmt.Errorf("gagal membuat QR code: %w", err)
	}

	// Simpan ke file
	png, err := qr.PNG(256)
	if err != nil {
		return fmt.Errorf("gagal menghasilkan PNG QR code: %w", err)
	}

	if err := ioutil.WriteFile(h.qrCodePath, png, 0644); err != nil {
		return fmt.Errorf("gagal menyimpan QR code ke file: %w", err)
	}

	// Simpan string QR code untuk debugging
	txtPath := filepath.Join(filepath.Dir(h.qrCodePath), "latest-qr.txt")
	if err := ioutil.WriteFile(txtPath, []byte(qrCode), 0644); err != nil {
		h.logger.WithError(err).Warn("Gagal menyimpan teks QR code")
	}

	// Update timestamp
	h.timestamp = time.Now()

	// Simpan timestamp ke file
	timestampPath := filepath.Join(filepath.Dir(h.qrCodePath), "qr-timestamp.txt")
	if err := ioutil.WriteFile(timestampPath, []byte(h.timestamp.Format(time.RFC3339)), 0644); err != nil {
		h.logger.WithError(err).Warn("Gagal menyimpan timestamp QR code")
	}

	h.logger.WithFields(utils.Fields{
		"path":      h.qrCodePath,
		"timestamp": h.timestamp,
	}).Info("QR code berhasil disimpan")

	return nil
}

// GetQRCodeImage mengembalikan QR code sebagai base64 string
func (h *QRHandler) GetQRCodeImage() (string, time.Time, error) {
	// Periksa apakah file QR code ada
	if _, err := os.Stat(h.qrCodePath); os.IsNotExist(err) {
		return "", time.Time{}, fmt.Errorf("QR code belum dibuat")
	}

	// Baca file QR code
	data, err := ioutil.ReadFile(h.qrCodePath)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("gagal membaca file QR code: %w", err)
	}

	// Konversi ke base64
	b64Data := base64.StdEncoding.EncodeToString(data)

	return b64Data, h.timestamp, nil
}

// GetQRCodeTimestamp mengembalikan waktu pembuatan QR code terakhir
func (h *QRHandler) GetQRCodeTimestamp() time.Time {
	// Jika timestamp sudah diset di memory, gunakan itu
	if !h.timestamp.IsZero() {
		return h.timestamp
	}

	// Jika tidak, coba baca dari file
	timestampPath := filepath.Join(filepath.Dir(h.qrCodePath), "qr-timestamp.txt")
	if _, err := os.Stat(timestampPath); os.IsNotExist(err) {
		return time.Time{}
	}

	data, err := ioutil.ReadFile(timestampPath)
	if err != nil {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}
	}

	h.timestamp = t
	return t
}

// IsQRCodeExpired memeriksa apakah QR code sudah kedaluwarsa
func (h *QRHandler) IsQRCodeExpired(maxAgeMinutes int) bool {
	timestamp := h.GetQRCodeTimestamp()
	if timestamp.IsZero() {
		return true
	}

	return time.Since(timestamp).Minutes() > float64(maxAgeMinutes)
}

// ClearQRCode menghapus QR code yang disimpan
func (h *QRHandler) ClearQRCode() {
	// Hapus file QR code
	if err := os.Remove(h.qrCodePath); err != nil && !os.IsNotExist(err) {
		h.logger.WithError(err).Warn("Gagal menghapus file QR code")
	}

	// Hapus file text QR code
	txtPath := filepath.Join(filepath.Dir(h.qrCodePath), "latest-qr.txt")
	if err := os.Remove(txtPath); err != nil && !os.IsNotExist(err) {
		h.logger.WithError(err).Warn("Gagal menghapus file text QR code")
	}

	// Reset timestamp
	h.timestamp = time.Time{}
}

// GetQRCodeData mengembalikan QR code data saat ini dan timestamp ketika dibuat
func (h *QRHandler) GetQRCodeData() (string, time.Time) {
	txtPath := filepath.Join(filepath.Dir(h.qrCodePath), "latest-qr.txt")
	if _, err := os.Stat(txtPath); os.IsNotExist(err) {
		// File QR code tidak ada
		return "", time.Time{}
	}

	// Baca QR code dari file
	qrData, err := ioutil.ReadFile(txtPath)
	if err != nil {
		h.logger.WithError(err).Error("Gagal membaca file QR code")
		return "", time.Time{}
	}

	// Dapatkan timestamp
	timestamp := h.GetQRCodeTimestamp()

	return string(qrData), timestamp
}

// ProcessQRCode menangani semua aspek pemrosesan QR code
func (h *QRHandler) ProcessQRCode(qrCode string) error {
	// Validasi QR code
	qrCode = strings.TrimSpace(qrCode)
	if qrCode == "" {
		return fmt.Errorf("QR code kosong")
	}

	// 1. Tampilkan QR code di terminal
	printQRToTerminal(qrCode)

	// 2. Simpan QR code ke file
	if err := h.SaveQRCode(qrCode); err != nil {
		return fmt.Errorf("gagal menyimpan QR code: %w", err)
	}

	// 3. Perbarui timestamp
	h.timestamp = time.Now()

	return nil
}

// GetQRCodePath mengembalikan path file QR code
func (h *QRHandler) GetQRCodePath() string {
	return h.qrCodePath
}

// printQRToTerminal menampilkan QR code di terminal sebagai ASCII art
func printQRToTerminal(qrCodeStr string) {

	qr, err := qrcode.New(qrCodeStr, qrcode.Medium)
	if err != nil {
		return
	}

	// Tampilkan QR code sebagai ASCII art di terminal
	fmt.Println(qr.ToSmallString(false))
}
