package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
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
	// Log path file template untuk debugging
	c.logger.WithFields(utils.Fields{
		"template": "index",
	}).Debug("Rendering template")

	return ctx.Render("index", fiber.Map{
		"Title":       "WhatsApp Bot Notify",
		"Description": "Aplikasi notifikasi WhatsApp",
		"Version":     "1.0.0",
	})
}
