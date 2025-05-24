package api

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes mendaftarkan semua endpoint API ke Fiber app
func (h *APIHandler) RegisterEndpoints(app *fiber.App) {
	// API health check tanpa autentikasi
	app.Get("/ping", h.statusHandler.TestConnection)

	// Grup API dengan autentikasi
	api := app.Group("/api")
	api.Use(h.authMw)

	// Status API
	api.Get("/status", h.statusHandler.GetStatus)

	// WhatsApp Connection
	api.Post("/reconnect", h.connHandler.Reconnect)
	api.Get("/reconnect", h.connHandler.Reconnect)
	api.Post("/disconnect", h.connHandler.Disconnect)

	// Message API
	api.Post("/send/personal", h.msgHandler.SendPersonal)
	api.Post("/send/group", h.msgHandler.SendGroup)

	// Groups API
	api.Get("/groups", h.groupHandler.ListGroups)

	// QR Code API
	api.Get("/qr/status", h.qrHandler.GetStatus)
	api.Get("/qr/image", h.qrHandler.GetImage)

	// Logs API
	api.Get("/logs", h.logsHandler.GetLogs)
	api.Post("/logs/clear", h.logsHandler.ClearLogs)
	api.Get("/logs/export", h.logsHandler.ExportLogs)
}
