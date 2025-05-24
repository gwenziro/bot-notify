package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/log"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
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
	storage      storage.Storage

	// Controller untuk berbagai halaman
	homeController         *controller.HomeController
	statusController       *controller.StatusController
	connectivityController *controller.ConnectivityController
	dashboardController    *controller.DashboardController
	settingsController     *controller.SettingsController
	authController         *controller.AuthController
	logsController         *controller.LogsController
}

// NewWebHandler membuat instance baru WebHandler
func NewWebHandler(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store) *WebHandler {
	// Sesuaikan path dengan struktur direktori baru
	viewsPath := filepath.Join(utils.ProjectRoot, "internal", "web", "view")
	staticPath := filepath.Join(utils.ProjectRoot, "static")

	logger := utils.ForModule("web")

	// Buat instance LogService
	// Catatan: Dalam produksi, ini sebaiknya diinjeksi dari luar
	logService := log.NewLogService(nil, utils.ForModule("log-service"))

	// Inisialisasi storage dengan fungsi konstruktor yang benar
	storageInstance, err := storage.NewStorage(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize storage: %v", err)
	}

	// Inisialisasi controller
	homeController := controller.NewHomeController(cfg, whatsClient, logger)
	statusController := controller.NewStatusController(cfg, whatsClient, logger)
	connectivityController := controller.NewConnectivityController(cfg, whatsClient, logger)
	dashboardController := controller.NewDashboardController(cfg, whatsClient, logger)
	settingsController := controller.NewSettingsController(cfg, whatsClient, logger)
	authController := controller.NewAuthController(cfg, whatsClient, sessionStore, storageInstance, logger)
	logsController := controller.NewLogsController(cfg, whatsClient, logService, logger)

	return &WebHandler{
		config:                 cfg,
		whatsApp:               whatsClient,
		logger:                 logger,
		viewsPath:              viewsPath,
		staticPath:             staticPath,
		sessionStore:           sessionStore,
		storage:                storageInstance,
		homeController:         homeController,
		statusController:       statusController,
		connectivityController: connectivityController,
		dashboardController:    dashboardController,
		settingsController:     settingsController,
		authController:         authController,
		logsController:         logsController,
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
	h.authController = controller.NewAuthController(h.config, h.whatsApp, store, h.storage, h.logger)
}
