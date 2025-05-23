package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthController menangani autentikasi halaman web
type AuthController struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	sessionStore *session.Store
}

// NewAuthController membuat instance baru AuthController
func NewAuthController(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store, logger utils.LogrusEntry) *AuthController {
	return &AuthController{
		config:       cfg,
		whatsApp:     whatsClient,
		logger:       logger.WithField("component", "auth-controller"),
		sessionStore: sessionStore,
	}
}

// LoginPage menampilkan halaman login
func (c *AuthController) LoginPage(ctx *fiber.Ctx) error {
	// Periksa apakah pengguna sudah login
	sess, err := c.sessionStore.Get(ctx)
	if err == nil {
		authToken := sess.Get("auth_token")
		if authToken != nil && authToken.(string) == c.config.Auth.AccessToken {
			// Redirect ke dashboard jika sudah login
			return ctx.Redirect("/dashboard")
		}
	}

	// Dapatkan parameter
	redirect := ctx.Query("redirect", "/dashboard")
	errorMsg := ctx.Query("error", "")

	var errorText string
	if errorMsg == "invalid_credentials" {
		errorText = "Token akses tidak valid. Silakan coba lagi."
	} else if errorMsg == "invalid_session" {
		errorText = "Sesi Anda telah kedaluwarsa. Silakan login kembali."
	}

	// Tambahkan log untuk debugging
	c.logger.Info("Rendering halaman login", utils.Fields{
		"redirect": redirect,
		"error":    errorMsg,
	})

	// Render halaman login - Template path diubah tanpa menggunakan layout
	return ctx.Render("auth/login", fiber.Map{
		"Title":       "Login - WhatsApp Bot Notify",
		"RedirectTo":  redirect,
		"Error":       errorText,
		"AccessToken": c.config.Auth.AccessToken[:3] + "..." + c.config.Auth.AccessToken[len(c.config.Auth.AccessToken)-3:],
	})
}

// ProcessLogin memproses login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Dapatkan token dari form
	token := ctx.FormValue("token")
	rememberMe := ctx.FormValue("remember_me") == "on"
	redirect := ctx.FormValue("redirect", "/dashboard")

	// Validasi token
	if token != c.config.Auth.AccessToken {
		c.logger.Warn("Percobaan login gagal", utils.Fields{
			"ip": ctx.IP(),
		})
		return ctx.Redirect("/login?error=invalid_credentials&redirect=" + redirect)
	}

	// Buat sesi
	sess, err := c.sessionStore.Get(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mendapatkan sesi saat login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan sesi")
	}

	// Set token dan status autentikasi di sesi
	sess.Set("auth_token", token)
	sess.Set("authenticated", true)

	if err := sess.Save(); err != nil {
		c.logger.WithError(err).Error("Gagal menyimpan sesi login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan saat menyimpan sesi")
	}

	// Set cookie auto-login jika "remember me" dicentang
	if rememberMe {
		ctx.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    token,
			Path:     "/",
			MaxAge:   c.config.Auth.CookieMaxAge,
			Secure:   c.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	}

	c.logger.Info("Login berhasil", utils.Fields{
		"ip":          ctx.IP(),
		"remember_me": rememberMe,
	})

	// Redirect ke halaman yang diminta
	return ctx.Redirect(redirect)
}

// Logout menangani logout pengguna
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Dapatkan dan hapus sesi
	sess, err := c.sessionStore.Get(ctx)
	if err == nil {
		sess.Delete("auth_token")
		sess.Delete("authenticated")
		if err := sess.Save(); err != nil {
			c.logger.WithError(err).Warn("Gagal menghapus sesi saat logout")
		}
	}

	// Hapus cookie auto-login
	ctx.Cookie(&fiber.Cookie{
		Name:     "auto_login",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   c.config.Server.BaseURL != "http://localhost:8080",
		HTTPOnly: true,
		SameSite: "Strict",
	})

	c.logger.Info("User logged out", utils.Fields{
		"ip": ctx.IP(),
	})

	// Redirect ke halaman login
	return ctx.Redirect("/login")
}
