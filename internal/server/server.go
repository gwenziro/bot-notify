package server

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/gwenziro/bot-notify/internal/api"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web"
)

// ServerOptions adalah opsi untuk pembuatan server
type ServerOptions struct {
	// Opsi umum
	Config *config.Config

	// Opsi untuk web rendering
	EnableTemplateEngine bool
	ViewsPath            string

	// Handler yang akan didaftarkan
	WebHandler *web.WebHandler
	APIHandler *api.APIHandler
}

// Server adalah wrapper untuk aplikasi Fiber
type Server struct {
	App          *fiber.App
	SessionStore *session.Store
}

// NewServer membuat instance baru server dengan opsi yang diberikan
func NewServer(opts ServerOptions) (*Server, error) {
	// Siapkan konfigurasi dasar fiber
	fiberConfig := fiber.Config{
		ReadTimeout:  opts.Config.Server.ReadTimeout,
		WriteTimeout: opts.Config.Server.WriteTimeout,
	}

	// Setup template engine jika diaktifkan
	if opts.EnableTemplateEngine && opts.ViewsPath != "" {
		// Pastikan direktori views ada
		if err := utils.EnsureDirectoryExists(opts.ViewsPath); err != nil {
			return nil, err
		}

		// Pastikan direktori layouts ada
		layoutDir := filepath.Join(opts.ViewsPath, "layouts")
		if err := utils.EnsureDirectoryExists(layoutDir); err != nil {
			return nil, err
		}

		// Setup template engine
		engine := html.New(opts.ViewsPath, ".html")

		// Tambahkan fungsi-fungsi helper untuk template
		engine.AddFunc("formatDate", func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04:05")
		})

		// Set engine ke konfigurasi fiber
		fiberConfig.Views = engine

		// Gunakan error handler untuk web jika template engine diaktifkan
		fiberConfig.ErrorHandler = createWebErrorHandler()
	} else {
		// Gunakan error handler API default jika template engine tidak diaktifkan
		fiberConfig.ErrorHandler = createAPIErrorHandler()
	}

	// Buat Fiber app dengan konfigurasi
	app := fiber.New(fiberConfig)

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

	// Setup session store
	sessionStore := session.New(session.Config{
		Expiration: 24 * time.Hour,
		KeyLookup:  "cookie:" + opts.Config.Auth.CookieName,
	})

	// Daftarkan routes jika handler tersedia
	if opts.WebHandler != nil {
		if opts.EnableTemplateEngine {
			// Serve static files jika template engine diaktifkan
			staticPath := filepath.Join(utils.ProjectRoot, "static")
			if err := utils.EnsureDirectoryExists(staticPath); err != nil {
				return nil, err
			}
			app.Static("/static", staticPath)
		}

		// Daftarkan web routes
		opts.WebHandler.RegisterRoutes(app)
	}

	if opts.APIHandler != nil {
		// Daftarkan API routes
		opts.APIHandler.RegisterEndpoints(app)
	}

	utils.Info("Server berhasil diinisialisasi", utils.Fields{
		"template_engine": opts.EnableTemplateEngine,
		"views_path":      opts.ViewsPath,
	})

	return &Server{
		App:          app,
		SessionStore: sessionStore,
	}, nil
}

// createAPIErrorHandler membuat handler kesalahan untuk API
func createAPIErrorHandler() func(*fiber.Ctx, error) error {
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

// createWebErrorHandler membuat handler kesalahan untuk web
func createWebErrorHandler() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log error
		utils.Error("Web error", utils.Fields{
			"error": err.Error(),
			"code":  code,
			"path":  c.Path(),
		})

		// Coba render error page jika possible
		if err := c.Render("error", fiber.Map{
			"Title":   "Error",
			"Message": err.Error(),
			"Code":    code,
		}); err != nil {
			// Fallback ke JSON jika render error gagal
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
				"code":    code,
			})
		}

		return nil
	}
}
