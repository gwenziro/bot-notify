package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/middleware"
)

// RegisterRoutes mendaftarkan semua endpoint API ke Fiber app
func (h *APIHandler) RegisterEndpoints(app *fiber.App) {
	// API health check tanpa autentikasi
	app.Get("/ping", h.statusHandler.TestConnection)

	// Grup API dengan autentikasi
	api := app.Group("/api")

	// Tambahkan middleware untuk logging API
	api.Use(middleware.APILoggerMiddleware())

	// Lalu middleware autentikasi
	api.Use(h.authMw)

	// Daftarkan endpoint berdasarkan domain
	h.registerStatusEndpoints(api)
	h.registerConnectionEndpoints(api)
	h.registerMessageEndpoints(api)
	h.registerGroupEndpoints(api)
	h.registerQRCodeEndpoints(api)
	h.registerProfileEndpoints(api)
	h.registerAdminEndpoints(api)
}

// registerStatusEndpoints mendaftarkan endpoint status
func (h *APIHandler) registerStatusEndpoints(api fiber.Router) {
	api.Get("/status", h.statusHandler.GetStatus)
}

// registerConnectionEndpoints mendaftarkan endpoint koneksi
func (h *APIHandler) registerConnectionEndpoints(api fiber.Router) {
	api.Post("/reconnect", h.connHandler.Reconnect)
	api.Get("/reconnect", h.connHandler.Reconnect)
	api.Post("/disconnect", h.connHandler.Disconnect)
}

// registerMessageEndpoints mendaftarkan endpoint pesan
func (h *APIHandler) registerMessageEndpoints(api fiber.Router) {
	send := api.Group("/send")
	send.Post("/personal", h.msgHandler.SendPersonal)
	send.Post("/group", h.msgHandler.SendGroup)
	send.Post("/broadcast", h.broadcastHandler.SendBroadcast)
}

// registerGroupEndpoints mendaftarkan endpoint grup
func (h *APIHandler) registerGroupEndpoints(api fiber.Router) {
	api.Get("/groups", h.groupHandler.ListGroups)
	api.Get("/groups/:id/participants", h.groupHandler.GetParticipants)
}

// registerQRCodeEndpoints mendaftarkan endpoint QR code
func (h *APIHandler) registerQRCodeEndpoints(api fiber.Router) {
	qr := api.Group("/qr")
	qr.Get("/status", h.qrHandler.GetStatus)
	qr.Get("/image", h.qrHandler.GetImage)
}

// registerProfileEndpoints mendaftarkan endpoint profil
func (h *APIHandler) registerProfileEndpoints(api fiber.Router) {
	api.Get("/profile", h.profileHandler.GetProfile)
}

// registerAdminEndpoints mendaftarkan endpoint khusus admin
func (h *APIHandler) registerAdminEndpoints(api fiber.Router) {
	admin := api.Group("/admin")
	
	// Endpoint untuk melihat status semua pengguna (hanya admin)
	admin.Get("/users/status", h.statusHandler.GetAllUsersStatus)
}