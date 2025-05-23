package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/controller"
)

// WebHandler menangani endpoint dan tampilan web
type WebHandler struct {
	config     *config.Config
	whatsApp   *client.Client
	logger     utils.LogrusEntry
	viewsPath  string
	staticPath string
	// Controller untuk berbagai halaman
	homeController   *controller.HomeController
	statusController *controller.StatusController
	qrCodeController *controller.QRCodeController
}

// NewWebHandler membuat instance baru WebHandler
func NewWebHandler(cfg *config.Config, whatsClient *client.Client) *WebHandler {
	// Sesuaikan path dengan struktur direktori baru
	viewsPath := filepath.Join(utils.ProjectRoot, "internal", "web", "view")
	staticPath := filepath.Join(utils.ProjectRoot, "static")

	logger := utils.ForModule("web")

	// Inisialisasi controller
	homeController := controller.NewHomeController(cfg, whatsClient, logger)
	statusController := controller.NewStatusController(cfg, whatsClient, logger)
	qrCodeController := controller.NewQRCodeController(cfg, whatsClient, logger)

	// Buat instance WebHandler
	return &WebHandler{
		config:           cfg,
		whatsApp:         whatsClient,
		logger:           logger,
		viewsPath:        viewsPath,
		staticPath:       staticPath,
		homeController:   homeController,
		statusController: statusController,
		qrCodeController: qrCodeController,
	}
}

// Setup menginisialisasi middleware dan engine template
func (h *WebHandler) Setup(app *fiber.App) error {
	// Pastikan direktori views dan static ada
	if err := utils.EnsureDirectoryExists(h.viewsPath); err != nil {
		return err
	}
	if err := utils.EnsureDirectoryExists(h.staticPath); err != nil {
		return err
	}

	// Serve static files
	app.Static("/static", h.staticPath)

	// Log info
	h.logger.Info("Web handler dikonfigurasi", utils.Fields{
		"views_path":  h.viewsPath,
		"static_path": h.staticPath,
	})

	return nil
}
