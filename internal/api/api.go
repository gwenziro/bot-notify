package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/handler"
	"github.com/gwenziro/bot-notify/internal/api/middleware"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/log"
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
}

// NewAPIHandler membuat instance baru API Handler
func NewAPIHandler(cfg *config.Config, whatsClient *client.Client, logService *log.LogService) *APIHandler {
	// Inisialisasi auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.AccessToken)

	// Inisialisasi handler-handler untuk setiap domain
	statusHandler := handler.NewStatusHandler(whatsClient)
	connHandler := handler.NewConnectionHandler(whatsClient)
	msgHandler := handler.NewMessageHandler(whatsClient)
	groupHandler := handler.NewGroupHandler(whatsClient)
	qrHandler := handler.NewQRCodeHandler(whatsClient)
	logsHandler := handler.NewLogsHandler(logService, utils.ForModule("api-logs"))

	return &APIHandler{
		statusHandler: statusHandler,
		connHandler:   connHandler,
		msgHandler:    msgHandler,
		groupHandler:  groupHandler,
		qrHandler:     qrHandler,
		logsHandler:   logsHandler,
		authMw:        authMiddleware.Validate,
	}
}
