package server

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// FiberApp adalah wrapper untuk aplikasi Fiber dengan session store
type FiberApp struct {
	App          *fiber.App
	SessionStore *session.Store
}

// NewFiberApp membuat instance baru dari aplikasi Fiber dengan konfigurasi
func NewFiberApp(cfg *config.Config) (*FiberApp, error) {
	// Buat Fiber app dengan config yang sesuai untuk API
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: createErrorHandler(),
	})

	// Add standard middleware
	app.Use(recover.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
		Output: os.Stdout,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, X-Access-Token, Authorization",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length, Content-Type",
	}))
	app.Use(compress.New())

	// Hapus static file serving karena kita fokus ke API

	// Setup session store dengan konfigurasi yang benar
	sessionStore := session.New(session.Config{
		Expiration: 24 * time.Hour,
		KeyLookup:  "cookie:" + cfg.Auth.CookieName,
	})

	return &FiberApp{
		App:          app,
		SessionStore: sessionStore,
	}, nil
}

// createErrorHandler membuat handler kesalahan kustom
func createErrorHandler() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log error
		utils.Error("API error", utils.Fields{
			"error": err.Error(),
			"code":  code,
			"path":  c.Path(),
		})

		// Kembalikan respons JSON untuk error
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"code":    code,
		})
	}
}
