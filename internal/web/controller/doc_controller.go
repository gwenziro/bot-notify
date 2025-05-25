package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// DocController menangani halaman dokumentasi
type DocController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewDocController membuat instance baru DocController
func NewDocController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DocController {
	return &DocController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "doc-controller"),
	}
}

// DocumentationPage menampilkan halaman dokumentasi API
func (c *DocController) DocumentationPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering documentation page")

	// Dapatkan status koneksi WhatsApp untuk sidebar
	connectionState := c.whatsApp.GetConnectionState()

	// Persiapkan contoh code dengan token yang disamarkan
	maskedToken := maskToken(c.config.Auth.AccessToken)

	// Persiapkan data untuk template
	baseURL := c.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Siapkan endpoints dengan method sudah dalam lowercase
	endpoints := []fiber.Map{
		{
			"Name":        "Status",
			"Endpoint":    "/api/status",
			"Method":      "GET",
			"MethodLower": "get", // Tambahkan ini
			"Description": "Mendapatkan status koneksi WhatsApp",
			"Example": `curl -X GET "{{.BaseURL}}/api/status" \\
  -H "X-Access-Token: {{.MaskedToken}}"`,
			"Response": `{
  "is_connected": true,
  "status": "connected",
  "device_name": "WhatsApp Web",
  "connected_since": "2025-05-25T23:15:20Z",
  "connection_retries": 0,
  "messages_sent": 5,
  "groups_count": 3
}`,
		},
		{
			"Name":        "Kirim Pesan Personal",
			"Endpoint":    "/api/send/personal",
			"Method":      "POST",
			"MethodLower": "post", // Tambahkan ini
			"Description": "Mengirim pesan ke nomor WhatsApp personal",
			"Example": `curl -X POST "{{.BaseURL}}/api/send/personal" \\
  -H "X-Access-Token: {{.MaskedToken}}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "phone": "628123456789",
    "message": "Ini adalah pesan notifikasi"
  }'`,
			"Response": `{
  "success": true,
  "message_id": "12345",
  "timestamp": "2025-05-25T23:20:15Z"
}`,
		},
		{
			"Name":        "Kirim Pesan Grup",
			"Endpoint":    "/api/send/group",
			"Method":      "POST",
			"MethodLower": "post", // Tambahkan ini
			"Description": "Mengirim pesan ke grup WhatsApp",
			"Example": `curl -X POST "{{.BaseURL}}/api/send/group" \\
  -H "X-Access-Token: {{.MaskedToken}}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "group_id": "120363123456789@g.us",
    "message": "Ini adalah pesan notifikasi grup"
  }'`,
			"Response": `{
  "success": true,
  "message_id": "67890",
  "timestamp": "2025-05-25T23:22:30Z"
}`,
		},
		{
			"Name":        "Daftar Grup",
			"Endpoint":    "/api/groups",
			"Method":      "GET",
			"MethodLower": "get", // Tambahkan ini
			"Description": "Mendapatkan daftar grup WhatsApp",
			"Example": `curl -X GET "{{.BaseURL}}/api/groups" \\
  -H "X-Access-Token: {{.MaskedToken}}"`,
			"Response": `{
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
	}

	// Buat data untuk template
	data := fiber.Map{
		"Title":             "Dokumentasi API",
		"CurrentYear":       time.Now().Year(),
		"BaseURL":           baseURL,
		"MaskedToken":       maskedToken,
		"ActivePage":        "docs",                      // Untuk highlight menu aktif di sidebar
		"WhatsAppConnected": connectionState.IsConnected, // Untuk status di sidebar
		"Endpoints":         endpoints,
	}

	// Render dokumentasi dengan data
	return ctx.Render("documentation", data)
}
