package web

import "github.com/gofiber/fiber/v2"

// RegisterAPIRoutes mendaftarkan semua rute API ke Fiber app
func (h *WebHandler) RegisterAPIRoutes(app *fiber.App) {
	// API routes
	api := app.Group("/api")

	// Status endpoints
	api.Get("/status", h.apiHandler.statusHandler.GetStatus)
	api.Get("/ping", h.apiHandler.statusHandler.TestConnection)

	// QR Code endpoints
	api.Get("/qr/status", h.apiHandler.qrCodeHandler.GetStatus)
	api.Get("/qr/image", h.apiHandler.qrCodeHandler.GetImage)
}
