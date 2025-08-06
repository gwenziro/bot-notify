package website

import (
	"strings"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/service/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// DocService menyediakan fungsionalitas untuk halaman dokumentasi dengan dukungan multi-user
type DocService struct {
	config      *config.Config
	userManager *manager.UserManager
	logger      utils.LogrusEntry
}

// NewDocService membuat instance service dokumentasi baru dengan dukungan multi-user
func NewDocService(cfg *config.Config, userManager *manager.UserManager, logger utils.LogrusEntry) *DocService {
	return &DocService{
		config:      cfg,
		userManager: userManager,
		logger:      logger.WithField("component", "docs-service"),
	}
}

// GetDocumentationData menyiapkan data untuk halaman dokumentasi
func (s *DocService) GetDocumentationData() entity.DocumentationData {
	// Dapatkan status koneksi WhatsApp untuk sidebar (menggunakan admin client)
	var connectionState client.ConnectionState
	adminClient, exists := s.userManager.GetUserClient("admin")
	if exists {
		connectionState, _ = adminClient.GetConnectionStateSafe()
	} else {
		// Default state jika admin client tidak ada
		connectionState = client.ConnectionState{
			IsConnected: false,
			Status:      client.StatusDisconnected,
		}
	}

	// Persiapkan token asli
	token := s.config.Auth.AccessToken

	// Persiapkan data untuk template
	baseURL := utils.CleanBaseURL(s.config.Server.BaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Proses terlebih dahulu semua endpoint untuk mengganti placeholder
	endpoints := s.GetApiEndpoints()

	// Ganti placeholder dalam contoh endpoint dan respons dengan nilai sebenarnya
	for i := range endpoints {
		// Ganti {{.Token}} dengan token asli
		endpoints[i].Example = strings.ReplaceAll(endpoints[i].Example, "{{.Token}}", token)
		endpoints[i].Response = strings.ReplaceAll(endpoints[i].Response, "{{.Token}}", token)

		// Ganti {{.BaseURL}} dengan baseURL
		endpoints[i].Example = strings.ReplaceAll(endpoints[i].Example, "{{.BaseURL}}", baseURL)
		endpoints[i].Response = strings.ReplaceAll(endpoints[i].Response, "{{.BaseURL}}", baseURL)
	}

	// Buat data untuk template
	return entity.DocumentationData{
		Title:             "Dokumentasi API",
		CurrentYear:       time.Now().Year(),
		BaseURL:           baseURL,
		MaskedToken:       token, // Gunakan token asli untuk dokumentasi
		Token:             token,
		ActivePage:        "docs",
		WhatsAppConnected: connectionState.IsConnected,
		Endpoints:         endpoints,
	}
}

// GetApiEndpoints mengembalikan daftar endpoint API untuk dokumentasi dengan informasi multi-user
func (s *DocService) GetApiEndpoints() []entity.ApiEndpoint {
	return []entity.ApiEndpoint{
		// STATUS ENDPOINTS
		{
			Name:        "Status Koneksi",
			Endpoint:    "/api/status",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan status koneksi WhatsApp untuk pengguna tertentu. Setiap pengguna memiliki instance bot yang terpisah.",
			Example: `curl -X GET "{{.BaseURL}}/api/status" \\
  -H "X-Access-Token: user123"`,
			Response: `{
  "success": true,
  "message": "WhatsApp terhubung",
  "status": "connected",
  "isConnected": true,
  "lastActivity": "15 Jun 2023 14:30:25"
}`,
		},
		{
			Name:        "Status Semua Pengguna (Admin)",
			Endpoint:    "/api/admin/users/status",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan status koneksi semua pengguna. Hanya tersedia untuk admin dengan token khusus.",
			Example: `curl -X GET "{{.BaseURL}}/api/admin/users/status" \\
  -H "X-Access-Token: {{.Token}}"`,
			Response: `{
  "success": true,
  "message": "Status semua pengguna berhasil diambil",
  "totalUsers": 3,
  "users": [
    {
      "userId": "user123",
      "isConnected": true,
      "createdAt": "2023-06-15T10:30:00Z",
      "lastActive": "2023-06-15T14:30:25Z"
    },
    {
      "userId": "user456",
      "isConnected": false,
      "createdAt": "2023-06-15T11:00:00Z",
      "lastActive": "2023-06-15T13:45:10Z"
    }
  ]
}`,
		},
		{
			Name:        "Health Check",
			Endpoint:    "/ping",
			Method:      "GET",
			MethodLower: "get",
			Description: "Memeriksa apakah API berfungsi, tanpa perlu autentikasi. Berguna untuk monitoring.",
			Example:     `curl -X GET "{{.BaseURL}}/ping"`,
			Response: `{
  "success": true,
  "message": "API berfungsi dengan baik",
  "version": "1.0.0"
}`,
		},

		// CONNECTION ENDPOINTS
		{
			Name:        "Reconnect WhatsApp",
			Endpoint:    "/api/reconnect",
			Method:      "POST",
			MethodLower: "post",
			Description: "Memulai ulang koneksi WhatsApp untuk pengguna tertentu. Jika belum ada sesi, akan memunculkan QR code baru. Setiap pengguna memiliki instance bot yang terpisah.",
			Example: `curl -X POST "{{.BaseURL}}/api/reconnect" \\
  -H "X-Access-Token: user123" \\
  -H "Content-Type: application/json"`,
			Response: `{
  "success": true,
  "message": "Proses reconnect berhasil dimulai",
  "status": "connecting",
  "isConnected": false
}`,
		},
		{
			Name:        "Disconnect WhatsApp",
			Endpoint:    "/api/disconnect",
			Method:      "POST",
			MethodLower: "post",
			Description: "Memutuskan koneksi WhatsApp untuk pengguna tertentu dan menghapus sesi. Memerlukan pindai QR ulang untuk koneksi berikutnya.",
			Example: `curl -X POST "{{.BaseURL}}/api/disconnect" \\
  -H "X-Access-Token: user123" \\
  -H "Content-Type: application/json"`,
			Response: `{
  "success": true,
  "message": "WhatsApp berhasil diputuskan",
  "status": "disconnected",
  "isConnected": false
}`,
		},

		// MESSAGE ENDPOINTS
		{
			Name:        "Kirim Pesan Personal",
			Endpoint:    "/api/send/personal",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke nomor WhatsApp personal menggunakan instance bot pengguna tertentu. Mendukung format nomor internasional maupun lokal.",
			Example: `curl -X POST "{{.BaseURL}}/api/send/personal" \\
  -H "X-Access-Token: user123" \\
  -H "Content-Type: application/json" \\
  -d '{
    "phoneNumber": "628123456789",
    "message": "Ini adalah pesan notifikasi"
  }'`,
			Response: `{
  "success": true,
  "message": "Pesan berhasil dikirim",
  "recipient": "08123456789",
  "type": "personal",
  "sentTime": "15 Jun 2023 14:35:20"
}`,
		},
		{
			Name:        "Kirim Pesan Grup",
			Endpoint:    "/api/send/group",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke grup WhatsApp menggunakan instance bot pengguna tertentu. ID grup bisa didapatkan dari endpoint /api/groups.",
			Example: `curl -X POST "{{.BaseURL}}/api/send/group" \\
  -H "X-Access-Token: user123" \\
  -H "Content-Type: application/json" \\
  -d '{
    "groupID": "120363123456789@g.us",
    "message": "Ini adalah pesan notifikasi grup"
  }'`,
			Response: `{
  "success": true,
  "message": "Pesan grup berhasil dikirim",
  "recipient": "120363123456789@g.us",
  "type": "group",
  "sentTime": "15 Jun 2023 14:40:10"
}`,
		},
		{
			Name:        "Broadcast Pesan",
			Endpoint:    "/api/send/broadcast",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke banyak penerima sekaligus (grup dan personal) menggunakan instance bot pengguna tertentu. Parameter delayMs mengatur jeda antar pengiriman dalam milidetik untuk mencegah throttling.",
			Example: `curl -X POST "{{.BaseURL}}/api/send/broadcast" \\
  -H "X-Access-Token: user123" \\
  -H "Content-Type: application/json" \\
  -d '{
    "personalNumbers": ["628123456789", "628987654321"],
    "groupIds": ["120363144182570365@g.us"],
    "message": "Ini adalah pesan broadcast",
    "delayMs": 1000
  }'`,
			Response: `{
  "success": true,
  "message": "Broadcast berhasil diproses",
  "results": [
    {
      "target": "628123456789",
      "type": "personal",
      "success": true,
      "sentTime": "15 Jun 2023 14:45:10"
    },
    {
      "target": "628987654321",
      "type": "personal",
      "success": true,
      "sentTime": "15 Jun 2023 14:45:11"
    },
    {
      "target": "120363144182570365@g.us",
      "type": "group",
      "success": true,
      "sentTime": "15 Jun 2023 14:45:12"
    }
  ],
  "successCount": 3,
  "failedCount": 0,
  "totalTargets": 3,
  "processingTimeMs": 2150,
  "sentTime": "15 Jun 2023 14:45:12"
}`,
		},

		// GROUP ENDPOINTS
		{
			Name:        "Daftar Grup",
			Endpoint:    "/api/groups",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan daftar grup WhatsApp yang diikuti oleh instance bot pengguna tertentu, termasuk informasi jumlah peserta dan status admin.",
			Example: `curl -X GET "{{.BaseURL}}/api/groups" \\
  -H "X-Access-Token: user123"`,
			Response: `{
  "success": true,
  "message": "Berhasil mendapatkan daftar grup",
  "groups": [
    {
      "id": "120363123456789@g.us",
      "name": "Grup Notifikasi",
      "participants": 25,
      "isAdmin": true,
      "createdAt": "2023-03-15T08:20:30Z"
    },
    {
      "id": "120363987654321@g.us",
      "name": "Tim Pengembang",
      "participants": 8,
      "isAdmin": false,
      "createdAt": "2023-02-10T14:25:10Z"
    }
  ],
  "total": 2
}`,
		},
		{
			Name:        "Daftar Anggota Grup",
			Endpoint:    "/api/groups/{id}/participants",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan daftar anggota dari grup WhatsApp tertentu menggunakan instance bot pengguna tertentu beserta informasi status admin mereka. Ganti {id} dengan ID grup.",
			Example: `curl -X GET "{{.BaseURL}}/api/groups/120363123456789@g.us/participants" \\
  -H "X-Access-Token: user123"`,
			Response: `{
  "success": true,
  "message": "Berhasil mendapatkan daftar peserta grup",
  "groupId": "120363123456789@g.us",
  "groupName": "Grup Notifikasi",
  "participants": [
    {
      "jid": "6281234567890@s.whatsapp.net",
      "displayName": "Admin Grup",
      "isAdmin": true
    },
    {
      "jid": "6289876543210@s.whatsapp.net",
      "displayName": "Anggota 1",
      "isAdmin": false
    }
  ],
  "total": 2,
  "adminCount": 1
}`,
		},

		// QR CODE ENDPOINTS
		{
			Name:        "Status QR Code",
			Endpoint:    "/api/qr/status",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan status QR code untuk proses koneksi WhatsApp pengguna tertentu. Endpoint ini membantu menentukan apakah QR code tersedia, kedaluwarsa, atau perlu diperbarui.",
			Example: `curl -X GET "{{.BaseURL}}/api/qr/status" \\
  -H "X-Access-Token: user123"`,
			Response: `{
  "success": true,
  "message": "QR code tersedia",
  "available": true,
  "expired": false,
  "timestamp": "2023-06-15T14:50:20Z",
  "qrUrl": "{{.BaseURL}}/api/qr/image?t=1686841820"
}`,
		},
		{
			Name:        "Gambar QR Code",
			Endpoint:    "/api/qr/image",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan gambar QR code untuk proses koneksi WhatsApp pengguna tertentu. Parameter t (timestamp) dapat ditambahkan untuk mencegah caching. Mengembalikan gambar PNG.",
			Example: `curl -X GET "{{.BaseURL}}/api/qr/image?t=1686841820" \\
  -H "X-Access-Token: user123" \\
  -o qrcode.png`,
			Response: `[Binary Image Data - PNG Format]`,
		},

		// PROFILE ENDPOINTS
		{
			Name:        "Profil WhatsApp",
			Endpoint:    "/api/profile",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan informasi profil akun WhatsApp yang terhubung untuk pengguna tertentu, termasuk nama, nomor telepon, dan foto profil jika tersedia.",
			Example: `curl -X GET "{{.BaseURL}}/api/profile" \\
  -H "X-Access-Token: user123"`,
			Response: `{
  "success": true,
  "message": "Berhasil mendapatkan profil",
  "profile": {
    "jid": "628123456789@s.whatsapp.net",
    "phoneNumber": "08123456789",
    "name": "Bot Notify",
    "pictureUrl": "https://pps.whatsapp.net/v/t61.24694-24/123456789_123456789_123456789_123456789_n.jpg?stp=dst-jpg_s96x96&ccb=11-4&oh=abc123&oe=ABC123",
    "status": "connected",
    "connectedSince": "15 Jun 2023 10:30:15",
    "lastActivity": "15 Jun 2023 14:55:30"
  }
}`,
		},
	}
}