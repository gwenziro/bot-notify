package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// StatusController menangani halaman status koneksi
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

// StatusPage menampilkan status koneksi WhatsApp
func (c *StatusController) StatusPage(ctx *fiber.Ctx) error {
	state := c.whatsApp.GetConnectionState()

	return ctx.Render("status", fiber.Map{
		"Title":          "Status Koneksi",
		"Status":         state.Status,
		"IsConnected":    state.IsConnected,
		"LastActivity":   state.LastActivity.Format("02 Jan 2006 15:04:05"),
		"ConnectedSince": state.Timestamp.Format("02 Jan 2006 15:04:05"),
		"Retries":        state.ConnectionRetries,
	})
}
