package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// DocController menangani halaman dokumentasi
type DocController struct {
	service *website.DocService
	logger  utils.LogrusEntry
}

// NewDocController membuat instance baru DocController
func NewDocController(service *website.DocService, logger utils.LogrusEntry) *DocController {
	return &DocController{
		service: service,
		logger:  logger.WithField("component", "doc-controller"),
	}
}

// DocumentationPage menampilkan halaman dokumentasi API
func (c *DocController) DocumentationPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering documentation page")

	// Persiapkan data untuk halaman dokumentasi melalui service
	docData := c.service.GetDocumentationData()

	// Render dokumentasi dengan data
	return ctx.Render("documentation", docData)
}
