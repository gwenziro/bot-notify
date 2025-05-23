package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// StatusController menangani halaman status
type StatusController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewStatusController membuat instance baru StatusController
func NewStatusController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *StatusController {
	return &StatusController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "status-controller"),
	}
}

// StatusPage menampilkan halaman status koneksi
func (c *StatusController) StatusPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman status")

	// Dapatkan status koneksi WhatsApp
	connectionState := c.whatsApp.GetConnectionState()

	return ctx.Render("dashboard/status", fiber.Map{
		"Title":        "Status Koneksi",
		"Description":  "Halaman untuk memantau status koneksi WhatsApp.",
		"Status":       string(connectionState.Status),
		"IsConnected":  connectionState.IsConnected,
		"LastActivity": connectionState.LastActivity,
		"Retries":      connectionState.ConnectionRetries,
		"DeviceInfo":   c.whatsApp.GetDeviceInfo(),
	})
}
