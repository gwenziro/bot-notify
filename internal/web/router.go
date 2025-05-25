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

	// Protected routes - Connectivity
	connectivity := app.Group("/connectivity")
	connectivity.Use(authMiddleware.RequireAuth())
	connectivity.Get("/", h.connectivityController.ConnectivityPage)

	// Protected routes - Status
	status := app.Group("/status")
	status.Use(authMiddleware.RequireAuth())
	status.Get("/", h.statusController.StatusPage)

	// Protected routes - Logs
	logs := app.Group("/logs")
	logs.Use(authMiddleware.RequireAuth())
	logs.Get("/", h.logsController.LogsPage)

	// Protected routes - Settings
	settings := app.Group("/settings")
	settings.Use(authMiddleware.RequireAuth())
	settings.Get("/", h.settingsController.SettingsPage)
	settings.Post("/update", h.settingsController.UpdateSettings)
	settings.Post("/token/update", h.settingsController.UpdateToken)
}
