package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/web/middleware"
)

// RegisterRoutes mendaftarkan semua rute web ke Fiber app
func (h *WebHandler) RegisterRoutes(app *fiber.App) {
	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.config, h.sessionStore)

	// Serve static files
	app.Static("/static", h.staticPath)

	// Public routes
	app.Get("/", h.homeController.HomePage)
	app.Get("/login", h.authController.LoginPage)
	app.Post("/auth/login", h.authController.ProcessLogin)
	app.Get("/logout", h.authController.Logout)

	// Protected routes - Dashboard
	dashboard := app.Group("/dashboard")
	dashboard.Use(authMiddleware.RequireAuth())
	dashboard.Get("/", h.dashboardController.DashboardPage)
	dashboard.Post("/refresh-qr", h.dashboardController.RefreshQRCode)
	dashboard.Post("/disconnect", h.dashboardController.DisconnectWhatsApp)

	// New Connectivity route
	connectivity := app.Group("/connectivity")
	connectivity.Use(authMiddleware.RequireAuth())

	// Documentation route - juga dilindungi auth
	docs := app.Group("/docs")
	docs.Use(authMiddleware.RequireAuth())
	docs.Get("/", h.docController.DocumentationPage)

	// API routes - tanpa middleware autentikasi khusus untuk sementara
	api := app.Group("/api")

	// Endpoint yang dibutuhkan oleh JavaScript dashboard
	api.Get("/status", h.apiHandler.statusHandler.GetStatus)
	api.Get("/qr/status", h.apiHandler.qrCodeHandler.GetStatus)
	api.Get("/qr/image", h.apiHandler.qrCodeHandler.GetImage)
}
