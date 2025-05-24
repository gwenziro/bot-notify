package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// connectivityController menangani halaman QR code
type ConnectivityController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewConnectivityController membuat instance baru ConnectivityController
func NewConnectivityController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *ConnectivityController {
	return &ConnectivityController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "qrcode-controller"),
	}
}

// ConnectivityPage menampilkan halaman QR code untuk koneksi WhatsApp
func (c *ConnectivityController) ConnectivityPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman QR code")

	// Dapatkan status koneksi
	connected := c.whatsApp.GetConnectionState().IsConnected

	// Dapatkan QR code jika tersedia
	var qrImage string
	var qrTimestamp time.Time
	var qrAvailable bool

	// Jika tidak terhubung, coba ambil QR code
	if !connected {
		qrHandler := c.whatsApp.SessionManager.GetQRHandler()
		if qrHandler != nil {
			image, timestamp, err := qrHandler.GetQRCodeImage()
			if err == nil && image != "" {
				qrImage = image
				qrTimestamp = timestamp
				qrAvailable = true
			}
		}
	}

	// Render dengan layout dashboard
	return ctx.Render("dashboard/connectivity", fiber.Map{
		"Title":            "Konektivitas WhatsApp",
		"Description":      "Halaman untuk menghubungkan WhatsApp Bot Notify.",
		"ActivePage":       "connectivity", // Untuk highlight menu aktif di sidebar
		"IsConnected":      connected,
		"QRCodeImage":      qrImage,
		"QRCodeTime":       qrTimestamp,
		"QRCodeAvailable":  qrAvailable,
		"RefreshInterval":  30, // Refresh setiap 30 detik
		"QRCodeEndpoint":   "/api/qr/image",
		"QRStatusEndpoint": "/api/qr/status",
	}, "layouts/dashboard" /* Gunakan layout dashboard */)
}
