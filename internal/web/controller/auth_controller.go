package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/auth"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/model"
)

// AuthController menangani autentikasi halaman web
type AuthController struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	sessionStore *session.Store
	authService  *auth.Service
}

// NewAuthController membuat instance baru AuthController
func NewAuthController(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store, logger utils.LogrusEntry) *AuthController {
	// Inisialisasi AuthController dengan auth service
	authService := auth.NewService(cfg, sessionStore, logger)

	return &AuthController{
		config:       cfg,
		whatsApp:     whatsClient,
		logger:       logger.WithField("component", "auth-controller"),
		sessionStore: sessionStore,
		authService:  authService,
	}
}

// LoginPage menampilkan halaman login
func (c *AuthController) LoginPage(ctx *fiber.Ctx) error {
	// Periksa apakah pengguna sudah login
	if c.authService.IsAuthenticated(ctx) {
		return ctx.Redirect("/dashboard")
	}

	// Dapatkan parameter
	redirect := ctx.Query("redirect", "/dashboard")
	errorMsg := ctx.Query("error", "")
	errorText := c.authService.GetErrorMessage(errorMsg)

	// Generate CSRF token
	csrfToken, err := c.authService.GenerateCSRFToken(ctx)
	if err != nil {
		c.logger.WithError(err).Warn("Gagal membuat CSRF token")
		// Lanjutkan meski gagal, halaman login masih bisa diakses
		return ctx.Redirect("/login?error=internal_error&redirect=" + redirect)
	}

	// Buat model login
	loginModel := model.NewLoginModel(redirect, errorText, csrfToken)
	loginModel.AccessToken = utils.MaskToken(c.config.Auth.AccessToken)

	// Tambahkan log untuk debugging
	c.logger.Info("Rendering halaman login", utils.Fields{
		"redirect": redirect,
		"error":    errorMsg,
	})

	// Render halaman login dengan model
	return ctx.Render("login", loginModel)
}

// ProcessLogin memproses login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Dapatkan redirect URL
	redirect := ctx.FormValue("redirect", "/dashboard")

	// Verifikasi CSRF token
	csrfToken := ctx.FormValue("csrf_token")
	if !c.authService.VerifyCSRFToken(ctx, csrfToken) {
		c.logger.Warn("CSRF token tidak valid", utils.Fields{
			"ip": ctx.IP(),
		})
		return ctx.Redirect("/login?error=invalid_request&redirect=" + redirect)
	}

	// Periksa rate limiting
	allowed, attempts, err := c.authService.CheckRateLimiting(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal memeriksa rate limiting")
		// Lanjutkan meskipun gagal
	}

	if !allowed {
		c.logger.Warn("Login rate limited", utils.Fields{
			"ip":       ctx.IP(),
			"attempts": attempts,
		})
		return ctx.Redirect("/login?error=too_many_attempts&redirect=" + redirect)
	}

	// Dapatkan token dan preferensi remember_me
	token := ctx.FormValue("token")
	rememberMe := ctx.FormValue("remember_me") == "on"

	// Validasi token
	if token != c.config.Auth.AccessToken {
		// Catat percobaan gagal
		if err := c.authService.IncrementLoginAttempt(ctx); err != nil {
			c.logger.WithError(err).Warn("Gagal mencatat percobaan login")
		}

		c.logger.Warn("Percobaan login gagal", utils.Fields{
			"ip": ctx.IP(),
		})
		return ctx.Redirect("/login?error=invalid_credentials&redirect=" + redirect)
	}

	// Token valid, reset percobaan
	if err := c.authService.ResetLoginAttempt(ctx); err != nil {
		c.logger.WithError(err).Warn("Gagal reset counter percobaan login")
	}

	// Buat sesi
	if err := c.authService.CreateSession(ctx, token, rememberMe); err != nil {
		c.logger.WithError(err).Error("Gagal membuat sesi login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan sesi")
	}

	c.logger.Info("Login berhasil", utils.Fields{
		"ip":          ctx.IP(),
		"remember_me": rememberMe,
	})

	// Redirect dengan localStorage
	redirectHTML := c.authService.GetRedirectHTML(token, redirect)
	ctx.Set("Content-Type", "text/html")
	return ctx.SendString(redirectHTML)
}

// Logout menangani logout pengguna
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Hapus sesi
	if err := c.authService.ClearSession(ctx); err != nil {
		c.logger.WithError(err).Warn("Gagal menghapus sesi saat logout")
	}

	c.logger.Info("User logged out", utils.Fields{
		"ip": ctx.IP(),
	})

	// Redirect ke halaman login
	return ctx.Redirect("/login")
}
