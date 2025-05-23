package web

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan semua rute web ke Fiber app
func (h *WebHandler) SetupRoutes(app *fiber.App) {
	// Serve static files
	app.Static("/static", h.staticPath)

	// Home page
	app.Get("/", h.homeController.HomePage)

	// Status page
	app.Get("/status", h.statusController.StatusPage)

	// QR Code page
	app.Get("/qr", h.qrCodeController.QRCodePage)
}
