package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/api/handler"
	"github.com/gwenziro/bot-notify/internal/api/middleware"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APIHandler bertanggung jawab untuk mengelola endpoint API
type APIHandler struct {
	// Handlers untuk berbagai domain
	statusHandler *handler.StatusHandler
	connHandler   *handler.ConnectionHandler
	msgHandler    *handler.MessageHandler
	groupHandler  *handler.GroupHandler
	qrHandler     *handler.QRCodeHandler
	logsHandler   *handler.LogsHandler
	authMw        fiber.Handler
	config        *config.Config
	whatsApp      *client.Client
	sessionStore  *session.Store
	logger        utils.LogrusEntry
}

// NewAPIHandler membuat instance baru APIHandler
func NewAPIHandler(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store) *APIHandler {
	logger := utils.ForModule("api")

	// Initialize API auth middleware (berbeda dengan web auth middleware)
	apiAuthMw := middleware.NewAPIAuthMiddleware(cfg, sessionStore)

	// Inisialisasi handler-handler untuk setiap domain
	statusHandler := handler.NewStatusHandler(whatsClient)
	connHandler := handler.NewConnectionHandler(whatsClient)
	msgHandler := handler.NewMessageHandler(whatsClient)
	groupHandler := handler.NewGroupHandler(whatsClient)
	qrHandler := handler.NewQRCodeHandler(whatsClient)
	// logsHandler := handler.NewLogsHandler(logService, utils.ForModule("api-logs"))

	return &APIHandler{
		statusHandler: statusHandler,
		connHandler:   connHandler,
		msgHandler:    msgHandler,
		groupHandler:  groupHandler,
		qrHandler:     qrHandler,
		// logsHandler:   logsHandler,
		authMw:       apiAuthMw.RequireAuth(),
		config:       cfg,
		whatsApp:     whatsClient,
		sessionStore: sessionStore,
		logger:       logger,
	}
}
