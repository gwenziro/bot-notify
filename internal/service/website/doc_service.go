package website

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// DocService menyediakan fungsionalitas untuk halaman dokumentasi
type DocService struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewDocService membuat instance service dokumentasi baru
func NewDocService(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DocService {
	return &DocService{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "docs-service"),
	}
}

// GetDocumentationData menyiapkan data untuk halaman dokumentasi
func (s *DocService) GetDocumentationData() entity.DocumentationData {
	// Dapatkan status koneksi WhatsApp untuk sidebar
	connectionState := s.whatsApp.GetConnectionState()

	// Persiapkan contoh code dengan token yang disamarkan
	maskedToken := utils.MaskToken(s.config.Auth.AccessToken)

	// Persiapkan data untuk template
	baseURL := s.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Buat data untuk template
	return entity.DocumentationData{
		Title:             "Dokumentasi API",
		CurrentYear:       time.Now().Year(),
		BaseURL:           baseURL,
		MaskedToken:       maskedToken,
		ActivePage:        "docs",
		WhatsAppConnected: connectionState.IsConnected,
		Endpoints:         s.GetApiEndpoints(),
	}
}

// GetApiEndpoints mengembalikan daftar endpoint API untuk dokumentasi
func (s *DocService) GetApiEndpoints() []entity.ApiEndpoint {
	return []entity.ApiEndpoint{
		{
			Name:        "Status",
			Endpoint:    "/api/status",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan status koneksi WhatsApp",
			Example: `curl -X GET "{{.BaseURL}}/api/status" \\
  -H "X-Access-Token: {{.MaskedToken}}"`,
			Response: `{
  "is_connected": true,
  "status": "connected",
  "device_name": "WhatsApp Web",
  "connected_since": "2025-05-25T23:15:20Z",
  "messages_sent": 5,
  "groups_count": 3
}`,
		},
		{
			Name:        "Kirim Pesan Personal",
			Endpoint:    "/api/send/personal",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke nomor WhatsApp personal",
			Example: `curl -X POST "{{.BaseURL}}/api/send/personal" \\
  -H "X-Access-Token: {{.MaskedToken}}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "phone": "628123456789",
    "message": "Ini adalah pesan notifikasi"
  }'`,
			Response: `{
  "success": true,
  "message_id": "12345",
  "timestamp": "2025-05-25T23:20:15Z"
}`,
		},
		{
			Name:        "Kirim Pesan Grup",
			Endpoint:    "/api/send/group",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke grup WhatsApp",
			Example: `curl -X POST "{{.BaseURL}}/api/send/group" \\
  -H "X-Access-Token: {{.MaskedToken}}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "group_id": "120363123456789@g.us",
    "message": "Ini adalah pesan notifikasi grup"
  }'`,
			Response: `{
  "success": true,
  "message_id": "67890",
  "timestamp": "2025-05-25T23:22:30Z"
}`,
		},
		{
			Name:        "Daftar Grup",
			Endpoint:    "/api/groups",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan daftar grup WhatsApp",
			Example: `curl -X GET "{{.BaseURL}}/api/groups" \\
  -H "X-Access-Token: {{.MaskedToken}}"`,
			Response: `{
  "groups": [
    {
      "id": "120363123456789@g.us",
      "name": "Grup Notifikasi",
      "participants_count": 25
    },
    {
      "id": "120363987654321@g.us",
      "name": "Tim Pengembang",
      "participants_count": 8
    }
  ]
}`,
		},
		{
			Name:        "Broadcast Pesan",
			Endpoint:    "/api/broadcast",
			Method:      "POST",
			MethodLower: "post",
			Description: "Mengirim pesan ke banyak tujuan sekaligus",
			Example: `curl -X POST "{{.BaseURL}}/api/broadcast" \\
  -H "X-Access-Token: {{.MaskedToken}}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "personalNumbers": ["628123456789", "628987654321"],
    "groupIds": ["120363144182570365@g.us"],
    "message": "Ini adalah pesan broadcast",
    "delayMs": 1000
  }'`,
			Response: `{
  "success": true,
  "total_targets": 3,
  "success_count": 3,
  "failed_count": 0,
  "processing_time_ms": 2150,
  "sent_time": "25 Mei 2025 23:25:30"
}`,
		},
	}
}
