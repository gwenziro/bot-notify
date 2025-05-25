package server

import (
	"context"
	"encoding/json"
	"fmt"
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
	// Siapkan konfigurasi dasar fiber dengan timeout yang lebih ketat
	fiberConfig := fiber.Config{
		ReadTimeout:           opts.Config.Server.ReadTimeout,
		WriteTimeout:          opts.Config.Server.WriteTimeout,
		IdleTimeout:           30 * time.Second, // Tambahkan idle timeout
		DisableStartupMessage: false,            // Aktifkan pesan startup
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

		// Setup template engine dengan debug info
		utils.Info("Mengonfigurasi template engine", utils.Fields{
			"views_path":  opts.ViewsPath,
			"layouts_dir": layoutDir,
		})

		engine := html.New(opts.ViewsPath, ".html")

		// Konfigurasi engine agar bekerja dengan layout
		engine.AddFunc("yield", func() string {
			return "{{embed}}"
		})

		engine.AddFunc("formatDate", func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04:05")
		})

		// Tambahkan fungsi currentYear untuk penggunaan di template
		engine.AddFunc("currentYear", func() string {
			return time.Now().Format("2006")
		})

		engine.AddFunc("json", func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		})

		engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires an even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		})

		// Reload templates untuk development
		engine.Reload(true)

		// Debug mode untuk lebih banyak informasi error
		engine.Debug(true)

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
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:" + opts.Config.Auth.CookieName,
		CookieSameSite: "Strict",
		CookieSecure:   opts.Config.Server.BaseURL != "http://localhost:8080",
		CookieHTTPOnly: true,
	})

	// Tambahkan middleware untuk mencegah request hanging
	app.Use(func(c *fiber.Ctx) error {
		// Set timeout konteks untuk mencegah request tergantung terlalu lama
		ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
		defer cancel()

		c.SetUserContext(ctx)
		return c.Next()
	})

	// Pass session store to WebHandler if it exists
	if opts.WebHandler != nil {
		// Set session store to WebHandler
		opts.WebHandler.SetSessionStore(sessionStore)

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

	utils.Info("Server berhasil diinisialisasi dengan timeout handling", utils.Fields{
		"template_engine": opts.EnableTemplateEngine,
		"views_path":      opts.ViewsPath,
		"read_timeout":    opts.Config.Server.ReadTimeout,
		"write_timeout":   opts.Config.Server.WriteTimeout,
		"idle_timeout":    30 * time.Second,
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

		// Jangan log error timeout untuk path tertentu
		if code == fiber.StatusRequestTimeout && (c.Path() == "/" || c.Path() == "/favicon.ico") {
			// Skip logging untuk timeout pada path yang sering di-poll
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   "Request timeout",
				"code":    code,
			})
		}

		// Log error lainnya seperti biasa
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
