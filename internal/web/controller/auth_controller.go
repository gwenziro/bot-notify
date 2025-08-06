package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthController menangani autentikasi halaman web
type AuthController struct {
	service *website.AuthService
	logger  utils.LogrusEntry
}

// NewAuthController membuat instance baru AuthController
func NewAuthController(service *website.AuthService, logger utils.LogrusEntry) *AuthController {
	return &AuthController{
		service: service,
		logger:  logger.WithField("component", "auth-controller"),
	}
}

// LoginPage menampilkan halaman login
func (c *AuthController) LoginPage(ctx *fiber.Ctx) error {
	// Dapatkan parameter dari request
	redirect := ctx.Query("redirect", "/dashboard")
	errorCode := ctx.Query("error", "")

	// Periksa apakah pengguna sudah login melalui service
	isLoggedIn, err := c.service.IsLoggedIn(ctx)
	if err == nil && isLoggedIn {
		return ctx.Redirect(redirect)
	}

	// Dapatkan data untuk halaman login dari service
	pageData, err := c.service.PrepareLoginPage(ctx, redirect, errorCode)
	if err != nil {
		c.logger.WithError(err).Error("Gagal menyiapkan halaman login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan saat mempersiapkan halaman login")
	}

	// Log info untuk debugging
	c.logger.Info("Rendering halaman login", utils.Fields{
		"redirect": redirect,
		"error":    errorCode,
	})

	// Render halaman login
	return ctx.Render("login", pageData)
}

// ProcessLogin memproses login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Ambil data mentah dari form
	token := ctx.FormValue("token")
	rememberMe := ctx.FormValue("remember_me") == "on"
	redirect := ctx.FormValue("redirect", "/dashboard")
	csrfToken := ctx.FormValue("csrf_token")

	// Proses login melalui service
	success, errorCode, err := c.service.ProcessLogin(ctx, token, rememberMe, redirect, csrfToken)
	if err != nil {
		c.logger.WithError(err).Error("Error saat memproses login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan saat login")
	}

	if !success {
		// Redirect ke halaman login dengan pesan error
		return ctx.Redirect("/login?error=" + errorCode + "&redirect=" + redirect)
	}

	// Login berhasil, buat respons redirect
	return c.service.CreateLoginResponse(ctx, token, redirect)
}

// Logout menangani logout pengguna
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Hapus sesi via service
	if err := c.service.HandleLogout(ctx); err != nil {
		c.logger.WithError(err).Warn("Gagal memproses logout")
		// Tetap lakukan redirect meskipun terjadi error
	}

	// Redirect ke halaman login
	return ctx.Redirect("/login")
}
