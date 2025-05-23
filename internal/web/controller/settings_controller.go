package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// SettingsController menangani halaman pengaturan
type SettingsController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewSettingsController membuat instance baru SettingsController
func NewSettingsController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *SettingsController {
	return &SettingsController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "settings-controller"),
	}
}

// SettingsPage menampilkan halaman pengaturan
func (c *SettingsController) SettingsPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Render halaman pengaturan")

	// Mendapatkan informasi untuk ditampilkan di halaman pengaturan
	connectionInfo := c.whatsApp.GetConnectionInfo()

	// Data untuk tampilan
	viewData := fiber.Map{
		"Title":          "Pengaturan - WhatsApp Bot Notify",
		"ConnectionInfo": connectionInfo,
		"Config": fiber.Map{
			"ServerHost":      c.config.Server.Host,
			"ServerPort":      c.config.Server.Port,
			"ServerBaseURL":   c.config.Server.BaseURL,
			"MaxRetry":        c.config.WhatsApp.MaxRetry,
			"RetryDelay":      c.config.WhatsApp.RetryDelay.Seconds(),
			"IdleTimeout":     c.config.WhatsApp.IdleTimeout.Minutes(),
			"TokenExpiry":     c.config.Auth.TokenExpiry.Hours(),
			"LoggingLevel":    c.config.Logging.Level,
			"LoggingMaxSize":  c.config.Logging.MaxSize,
			"LoggingMaxAge":   c.config.Logging.MaxAge,
			"StorageType":     c.config.Storage.Type,
			"StorageInMemory": c.config.Storage.InMemory,
		},
		"Paths": fiber.Map{
			"StoreDir":    c.config.WhatsApp.StoreDir,
			"QrCodeDir":   c.config.WhatsApp.QrCodeDir,
			"SessionDir":  c.config.Auth.SessionDir,
			"StoragePath": c.config.Storage.Path,
			"LogFile":     c.config.Logging.File,
		},
	}

	// Render halaman
	return ctx.Render("dashboard/settings", viewData)
}

// UpdateSettings menerima pembaruan pengaturan dari form
func (c *SettingsController) UpdateSettings(ctx *fiber.Ctx) error {
	c.logger.Info("Menerima permintaan pembaruan pengaturan")

	// Di sini implementasi untuk memproses form pengaturan
	// Untuk sekarang, kita hanya redirect kembali dengan pesan sukses

	return ctx.Redirect("/settings?updated=true")
}

// UpdateToken memperbarui token akses API
func (c *SettingsController) UpdateToken(ctx *fiber.Ctx) error {
	c.logger.Info("Menerima permintaan pembaruan token API")

	// Di sini implementasi untuk regenerate token API
	// Untuk sekarang, kita hanya redirect kembali

	return ctx.Redirect("/settings?token_updated=true")
}
