package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// HomeController menangani halaman beranda web
type HomeController struct {
	service *website.HomeService
	logger  utils.LogrusEntry
}

// NewHomeController membuat instance baru HomeController
func NewHomeController(service *website.HomeService, logger utils.LogrusEntry) *HomeController {
	return &HomeController{
		service: service,
		logger:  logger.WithField("component", "home-controller"),
	}
}

// HomePage menampilkan halaman utama
func (c *HomeController) HomePage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman beranda")

	// Dapatkan data dari service
	pageData := c.service.GetHomePageData()

	// Render template dengan data
	return ctx.Render("index", pageData)
}
