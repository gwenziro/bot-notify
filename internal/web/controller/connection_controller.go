package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ConnectionController menangani koneksi WhatsApp dari web UI
type ConnectionController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewConnectionController membuat instance baru ConnectionController
func NewConnectionController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *ConnectionController {
	return &ConnectionController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "connection-controller"),
	}
}

// ReconnectWhatsApp menangani permintaan koneksi ulang dari web UI
func (c *ConnectionController) ReconnectWhatsApp(ctx *fiber.Ctx) error {
	c.logger.Info("Menangani permintaan reconnect dari web UI")

	// Inisiasi koneksi ulang
	err := c.whatsApp.Connect()
	if err != nil {
		c.logger.WithError(err).Error("Gagal menginisiasi reconnect")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memulai koneksi: " + err.Error(),
		})
	}

	// Redirect kembali ke dashboard
	return ctx.Redirect("/dashboard")
}
