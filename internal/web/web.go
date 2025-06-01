package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/controller"
)

// WebHandler menangani endpoint dan tampilan web
type WebHandler struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	viewsPath    string
	staticPath   string
	sessionStore *session.Store

	// Controller untuk berbagai halaman
	homeController       *controller.HomeController
	dashboardController  *controller.DashboardController
	authController       *controller.AuthController
	docController        *controller.DocController
	connectionController *controller.ConnectionController
}

// NewWebHandler membuat instance baru WebHandler
func NewWebHandler(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store) *WebHandler {
	// Sesuaikan path dengan struktur direktori baru
	viewsPath := filepath.Join(utils.ProjectRoot, "internal", "web", "view")
	staticPath := filepath.Join(utils.ProjectRoot, "static")

	logger := utils.ForModule("web")

	// Buat instance StatisticsService untuk dashboard
	statsService := website.NewStatisticsService()

	// Inisialisasi controller
	homeController := controller.NewHomeController(cfg, whatsClient, logger)
	dashboardController := controller.NewDashboardController(cfg, whatsClient, statsService, logger)
	authController := controller.NewAuthController(cfg, whatsClient, sessionStore, logger)
	docController := controller.NewDocController(cfg, whatsClient, logger)
	connectionController := controller.NewConnectionController(cfg, whatsClient, logger)

	return &WebHandler{
		config:               cfg,
		whatsApp:             whatsClient,
		logger:               logger,
		viewsPath:            viewsPath,
		staticPath:           staticPath,
		sessionStore:         sessionStore,
		homeController:       homeController,
		dashboardController:  dashboardController,
		authController:       authController,
		docController:        docController,
		connectionController: connectionController,
	}
}

// GetSessionStore mengembalikan session store
func (h *WebHandler) GetSessionStore() *session.Store {
	return h.sessionStore
}

// GetViewsPath mengembalikan path direktori template
func (h *WebHandler) GetViewsPath() string {
	return h.viewsPath
}

// GetStaticPath mengembalikan path direktori aset statis
func (h *WebHandler) GetStaticPath() string {
	return h.staticPath
}

// SetSessionStore menetapkan session store untuk WebHandler
func (h *WebHandler) SetSessionStore(store *session.Store) {
	h.sessionStore = store

	// Re-initialize auth controller with the new session store
	h.authController = controller.NewAuthController(h.config, h.whatsApp, store, h.logger)
}
