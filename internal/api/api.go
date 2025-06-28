package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/api/handler"
	"github.com/gwenziro/bot-notify/internal/api/middleware"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APIHandler bertanggung jawab untuk mengelola endpoint API dengan dukungan multi-user
type APIHandler struct {
	// Handlers untuk berbagai domain
	statusHandler    *handler.StatusHandler
	connHandler      *handler.ConnectionHandler
	msgHandler       *handler.MessageHandler
	broadcastHandler *handler.BroadcastHandler
	groupHandler     *handler.GroupHandler
	qrHandler        *handler.QRCodeHandler
	profileHandler   *handler.ProfileHandler
	authMw           fiber.Handler
	config           *config.Config
	userManager      *manager.UserManager
	sessionStore     *session.Store
	logger           utils.LogrusEntry
}

// NewAPIHandler membuat instance baru APIHandler dengan dukungan multi-user
func NewAPIHandler(cfg *config.Config, userManager *manager.UserManager, sessionStore *session.Store) *APIHandler {
	logger := utils.ForModule("api")

	// Lakukan validasi parameter
	if cfg == nil {
		logger.Error("Config tidak boleh nil saat membuat APIHandler")
		// Buat config default untuk mencegah crash
		cfg = &config.Config{}
	}

	if userManager == nil {
		logger.Error("UserManager nil saat membuat APIHandler")
	}

	// Initialize API auth middleware (berbeda dengan web auth middleware)
	apiAuthMw := middleware.NewAPIAuthMiddleware(cfg, sessionStore)

	// Inisialisasi handler-handler untuk setiap domain dengan UserManager
	statusHandler := handler.NewStatusHandler(userManager)
	connHandler := handler.NewConnectionHandler(userManager)
	msgHandler := handler.NewMessageHandler(userManager)
	broadcastHandler := handler.NewBroadcastHandler(userManager)
	groupHandler := handler.NewGroupHandler(userManager)
	qrHandler := handler.NewQRCodeHandler(userManager)
	profileHandler := handler.NewProfileHandler(userManager)

	return &APIHandler{
		statusHandler:    statusHandler,
		connHandler:      connHandler,
		msgHandler:       msgHandler,
		broadcastHandler: broadcastHandler,
		groupHandler:     groupHandler,
		qrHandler:        qrHandler,
		profileHandler:   profileHandler,
		authMw:           apiAuthMw.RequireAuth(),
		config:           cfg,
		userManager:      userManager,
		sessionStore:     sessionStore,
		logger:           logger,
	}
}