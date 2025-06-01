package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/model"
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

	// Buat model home
	homeModel := model.NewHomeModel(constants.DefaultAppVersion)
	homeModel.Description = "Aplikasi notifikasi WhatsApp"

	// Render template dengan model yang disederhanakan
	return ctx.Render("index", homeModel)
}
