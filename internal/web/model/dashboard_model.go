package model

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
)

// DashboardModel berisi data yang diperlukan untuk halaman dashboard
type DashboardModel struct {
	BasePageModel
	ConnectionState         client.ConnectionState // Status koneksi WhatsApp
	PhoneNumber             string                 // Nomor WhatsApp
	ContactName             string                 // Nama kontak WhatsApp
	MessagesSent            int                    // Jumlah pesan terkirim
	Uptime                  string                 // Waktu aplikasi berjalan
	BaseURL                 string                 // URL dasar aplikasi
	APIToken                string                 // Token untuk API calls (tidak dimasking)
	MaskedToken             string                 // Token API yang disamarkan untuk display
	ConnectedSince          time.Time              // Waktu koneksi dimulai (raw)
	ConnectedSinceFormatted string                 // Waktu koneksi dalam format yang mudah dibaca
	ConnectedDuration       string                 // Durasi koneksi dalam format yang mudah dibaca
	LastActivity            time.Time              // Waktu aktivitas terakhir (raw)
	LastActivityFormatted   string                 // Waktu aktivitas terakhir dalam format mudah dibaca

	// Tambahan untuk QR code
	QRCodeAvailable     bool   // Apakah QR code tersedia
	QRCodeExpired       bool   // Apakah QR code kedaluwarsa
	QRCodePath          string // Path ke gambar QR code
	QRCodeMessage       string // Pesan status QR code
	ShowReconnectButton bool   // Apakah perlu menampilkan tombol reconnect
}

// NewDashboardModel membuat instance baru DashboardModel
func NewDashboardModel(connectionState client.ConnectionState, connectedDuration string) DashboardModel {
	model := DashboardModel{
		BasePageModel:     NewBasePageModel("Dashboard", "dashboard"),
		ConnectionState:   connectionState,
		ConnectedDuration: connectedDuration,
	}
	model.IsConnected = connectionState.IsConnected
	return model
}
