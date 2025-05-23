package api

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes mendaftarkan semua rute API ke Fiber app
func (h *APIHandler) RegisterRoutes(app *fiber.App) {
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
}
