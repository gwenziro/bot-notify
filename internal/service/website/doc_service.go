package website

import (
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/model"
)

// Service mengelola dokumentasi API
type DocumentationService struct {
	baseURL     string
	maskedToken string
	logger      utils.LogrusEntry
}

// NewDocumentationService membuat instance baru DocumentationService
func NewDocumentationService(baseURL string, maskedToken string, logger utils.LogrusEntry) *DocumentationService {
	return &DocumentationService{
		baseURL:     baseURL,
		maskedToken: maskedToken,
		logger:      logger.WithField("component", "doc-service"),
	}
}

// GetAPIEndpoints mengembalikan daftar informasi endpoint API
func (s *DocumentationService) GetAPIEndpoints() []model.EndpointInfo {
	s.logger.Debug("Menyiapkan data dokumentasi endpoint API")

	return []model.EndpointInfo{
		{
			Name:        "Status",
			Endpoint:    "/api/status",
			Method:      "GET",
			MethodLower: "get",
			Description: "Mendapatkan status koneksi WhatsApp",
			Example: `curl -X GET "{{.BaseURL}}/api/status" \\
  -H "X-Access-Token: {{.MaskedToken}}"`,
			Response: `{
  "success": true,
  "message": "WhatsApp terhubung dan siap digunakan",
  "status": "connected",
  "is_connected": true,
  "connection_retries": 0,
  "last_activity": "27 Mei 2025 08:03:10"
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
    "phoneNumber": "628123456789",
    "message": "Ini adalah pesan notifikasi"
  }'`,
			Response: `{
  "success": true,
  "message": "Pesan WhatsApp terkirim!",
  "recipient": "628123456789@s.whatsapp.net",
  "type": "personal",
  "sentTime": "27 Mei 2025 08:03:10"
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
    "groupID": "120363123456789@g.us",
    "message": "Ini adalah pesan notifikasi grup"
  }'`,
			Response: `{
  "success": true,
  "message": "Pesan WhatsApp terkirim ke grup!",
  "recipient": "120363123456789@g.us",
  "type": "group",
  "sentTime": "27 Mei 2025 08:05:22"
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
  "success": true,
  "message": "Daftar grup berhasil diambil",
  "count": 2,
  "groups": [
    {
      "id": "120363123456789@g.us",
      "name": "Grup Notifikasi",
      "memberCount": 25,
      "isAdmin": true
    },
    {
      "id": "120363987654321@g.us",
      "name": "Tim Pengembang",
      "memberCount": 8,
      "isAdmin": false
    }
  ]
}`,
		},
	}
}

func (s *DocumentationService) GetBaseURL() string {
	return s.baseURL
}

func (s *DocumentationService) GetMaskedToken() string {
	return s.maskedToken
}
