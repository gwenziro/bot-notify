package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// DashboardController menangani halaman dashboard
type DashboardController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewDashboardController membuat instance baru DashboardController
func NewDashboardController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DashboardController {
	return &DashboardController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "dashboard-controller"),
	}
}

// DashboardPage menampilkan halaman dashboard utama
func (c *DashboardController) DashboardPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering dashboard page")

	// Dapatkan status koneksi
	connectionState := c.whatsApp.GetConnectionState()

	// Render dashboard dengan data
	return ctx.Render("dashboard/index", fiber.Map{
		"Title":       "Dashboard",
		"Description": "Ringkasan sistem dan status WhatsApp Bot Notify.",
		"Connection": fiber.Map{
			"Status":      string(connectionState.Status),
			"IsConnected": connectionState.IsConnected,
			"LastActive":  connectionState.LastActivity,
			"Retries":     connectionState.ConnectionRetries,
		},
		"SystemInfo": fiber.Map{
			"Version": "1.0.0",
			"Uptime":  "Loading...",
		},
	}, "layouts/dashboard")
}
