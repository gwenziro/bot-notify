package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/model"
)

// DocController menangani halaman dokumentasi
type DocController struct {
	config     *config.Config
	whatsApp   *client.Client
	logger     utils.LogrusEntry
	docService *website.DocumentationService
}

// NewDocController membuat instance baru DocController
func NewDocController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DocController {
	// Persiapkan data untuk dokumentasi
	baseURL := cfg.Server.BaseURL

	maskedToken := utils.MaskToken(cfg.Auth.AccessToken)

	// Buat service dokumentasi
	docService := website.NewDocumentationService(baseURL, maskedToken, logger)

	return &DocController{
		config:     cfg,
		whatsApp:   whatsClient,
		logger:     logger.WithField("component", "doc-controller"),
		docService: docService,
	}
}

// DocumentationPage menampilkan halaman dokumentasi API
func (c *DocController) DocumentationPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering documentation page")

	// Dapatkan status koneksi WhatsApp untuk sidebar
	connectionState := c.whatsApp.GetConnectionState()

	endpoints := c.docService.GetAPIEndpoints()
	docModel := model.NewDocumentationModel(c.docService.GetBaseURL(), c.docService.GetMaskedToken(), endpoints)
	docModel.IsConnected = connectionState.IsConnected

	// Render dokumentasi dengan model
	return ctx.Render("documentation", docModel)
}
