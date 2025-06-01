package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/api/handler"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/controller"
)

// APIHandler berisi semua handler API
type APIHandler struct {
	statusHandler  *handler.StatusHandler
	qrCodeHandler  *handler.QRCodeHandler
	groupHandler   *handler.GroupHandler
	messageHandler *handler.MessageHandler
}

// WebHandler menangani endpoint dan tampilan web
type WebHandler struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	viewsPath    string
	staticPath   string
	sessionStore *session.Store

	// Controller untuk berbagai halaman
	homeController      *controller.HomeController
	dashboardController *controller.DashboardController
	authController      *controller.AuthController
	docController       *controller.DocController

	// API handlers
	apiHandler *APIHandler
}

// NewWebHandler membuat instance baru WebHandler
func NewWebHandler(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store) *WebHandler {
	// Sesuaikan path dengan struktur direktori baru
	viewsPath := filepath.Join(utils.ProjectRoot, "internal", "web", "view")
	staticPath := filepath.Join(utils.ProjectRoot, "static")

	logger := utils.ForModule("web")

	// Inisialisasi controller
	homeController := controller.NewHomeController(cfg, whatsClient, logger)
	dashboardController := controller.NewDashboardController(cfg, whatsClient, logger)
	authController := controller.NewAuthController(cfg, whatsClient, sessionStore, logger)
	docController := controller.NewDocController(cfg, whatsClient, logger)

	// Inisialisasi API handler
	apiHandler := &APIHandler{
		statusHandler:  handler.NewStatusHandler(whatsClient),
		qrCodeHandler:  handler.NewQRCodeHandler(whatsClient),
		groupHandler:   handler.NewGroupHandler(whatsClient),
		messageHandler: handler.NewMessageHandler(whatsClient),
	}

	return &WebHandler{
		config:              cfg,
		whatsApp:            whatsClient,
		logger:              logger,
		viewsPath:           viewsPath,
		staticPath:          staticPath,
		sessionStore:        sessionStore,
		homeController:      homeController,
		dashboardController: dashboardController,
		authController:      authController,
		docController:       docController,
		apiHandler:          apiHandler,
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
