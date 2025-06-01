package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// HomeController menangani halaman beranda web
type HomeController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewHomeController membuat instance baru HomeController
func NewHomeController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *HomeController {
	return &HomeController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "home-controller"),
	}
}

// HomePage menampilkan halaman utama
func (c *HomeController) HomePage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman beranda")

	// Persiapkan data untuk halaman beranda
	pageData := entity.HomePageData{
		Title:       "WhatsApp Bot Notify",
		Description: "Bot WhatsApp Kirim Pesan Realtime",
		Version:     "1.0.0",
	}

	// Render template dengan data
	return ctx.Render("index", pageData)
}
