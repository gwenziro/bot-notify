package entity

import "time"

// DashboardData merepresentasikan data yang ditampilkan di halaman dashboard
type DashboardData struct {
	// Data Umum
	Title       string // Judul halaman
	CurrentYear int    // Tahun saat ini untuk footer
	BaseURL     string // URL dasar aplikasi
	MaskedToken string // Token API yang dimaskir (akan diganti dengan Token)
	Token       string // Token API yang ditampilkan penuh
	ActivePage  string // Halaman aktif untuk navigasi

	// Status WhatsApp
	IsConnected             bool      // Status koneksi WhatsApp
	ConnectionStatus        string    // Status koneksi sebagai string
	ConnectionDuration      string    // Durasi terhubung sebagai string
	ConnectedSince          time.Time // Waktu mulai terhubung
	ConnectedSinceFormatted string    // Waktu mulai terhubung yang sudah diformat
	ConnectedSinceShort     string    // Format waktu yang lebih singkat untuk card detail
	LastActivity            time.Time // Waktu aktivitas terakhir
	LastActivityFormatted   string    // Waktu aktivitas terakhir yang sudah diformat

	// Informasi Profil WhatsApp
	PhoneNumber       string // Nomor telepon yang terhubung
	ContactName       string // Nama kontak yang terhubung
	ProfilePictureURL string // URL foto profil WhatsApp

	// Statistik
	MessagesSent int // Jumlah pesan yang telah dikirim

	// Status QR Code
	QRCodeAvailable bool   // Status ketersediaan QR code
	QRCodeExpired   bool   // Status kedaluwarsa QR code
	QRCodeURL       string // URL QR code untuk ditampilkan
	QRCodeMessage   string // Pesan terkait QR code
}

// NewDashboardData membuat instance baru DashboardData dengan nilai default
func NewDashboardData(baseURL, token string) DashboardData {
	return DashboardData{
		Title:        "Dashboard - Bot Notify",
		CurrentYear:  time.Now().Year(),
		BaseURL:      baseURL,
		MaskedToken:  token,
		Token:        token,
		ActivePage:   "dashboard",
		MessagesSent: 0,
		IsConnected:  false,
	}
}
