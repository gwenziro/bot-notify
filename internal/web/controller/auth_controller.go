package controller

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthController menangani autentikasi halaman web
type AuthController struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	sessionStore *session.Store
	store        storage.Storage
}

// NewAuthController membuat instance baru AuthController
func NewAuthController(cfg *config.Config, whatsClient *client.Client, sessionStore *session.Store, logger utils.LogrusEntry) *AuthController {
	// Inisialisasi AuthController
	return &AuthController{
		config:       cfg,
		whatsApp:     whatsClient,
		logger:       logger.WithField("component", "auth-controller"),
		sessionStore: sessionStore,
		store:        storage.GetStorage(),
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
	} else if errorMsg == "too_many_attempts" {
		errorText = "Terlalu banyak percobaan login gagal. Silakan coba lagi nanti."
	} else if errorMsg == "session_timeout" {
		errorText = "Sesi Anda telah kedaluwarsa karena tidak aktif terlalu lama."
	} else if errorMsg == "security_concern" {
		errorText = "Terdapat masalah keamanan dengan sesi Anda. Silakan login kembali."
	} else if errorMsg == "invalid_request" {
		errorText = "Permintaan tidak valid. Silakan coba lagi."
	}

	// Generate CSRF token dan simpan dalam sesi
	csrfToken := generateRandomToken(32)
	sess, err = c.sessionStore.Get(ctx)
	if err == nil {
		sess.Set("csrf_token", csrfToken)
		if err := sess.Save(); err != nil {
			c.logger.WithError(err).Warn("Gagal menyimpan CSRF token ke sesi")
		}
	}

	// Tambahkan log untuk debugging
	c.logger.Info("Rendering halaman login", utils.Fields{
		"redirect": redirect,
		"error":    errorMsg,
	})

	// Render halaman login
	return ctx.Render("auth/login", fiber.Map{
		"Title":       "Login - WhatsApp Bot Notify",
		"RedirectTo":  redirect,
		"Error":       errorText,
		"CsrfToken":   csrfToken,
		"AccessToken": c.maskToken(c.config.Auth.AccessToken),
	})
}

// ProcessLogin memproses login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Verifikasi CSRF token
	csrfToken := ctx.FormValue("csrf_token")
	sess, err := c.sessionStore.Get(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mendapatkan sesi saat validasi CSRF")
		return ctx.Redirect("/login?error=invalid_request&redirect=" + ctx.FormValue("redirect", "/dashboard"))
	}

	storedToken := sess.Get("csrf_token")
	if storedToken == nil || csrfToken == "" || csrfToken != storedToken {
		c.logger.Warn("CSRF token tidak valid", utils.Fields{
			"ip": ctx.IP(),
		})
		return ctx.Redirect("/login?error=invalid_request&redirect=" + ctx.FormValue("redirect", "/dashboard"))
	}

	// Hapus CSRF token setelah digunakan
	sess.Delete("csrf_token")
	sess.Save()

	// Implementasi rate limiting sederhana dengan IP
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background() // Buat context untuk operasi storage

	attemptData, err := c.store.Get(bgCtx, ipKey)

	attempts := 0
	if err == nil && len(attemptData) > 0 {
		attempts, _ = strconv.Atoi(string(attemptData))

		// Jika melebihi 5 percobaan dalam 15 menit, tolak
		if attempts >= 5 {
			c.logger.Warn("Login rate limited", utils.Fields{
				"ip":       ctx.IP(),
				"attempts": attempts,
			})
			return ctx.Redirect("/login?error=too_many_attempts&redirect=" + ctx.FormValue("redirect", "/dashboard"))
		}
	}

	// Dapatkan token dari form
	token := ctx.FormValue("token")
	rememberMe := ctx.FormValue("remember_me") == "on"
	redirect := ctx.FormValue("redirect", "/dashboard")

	// Validasi token
	if token != c.config.Auth.AccessToken {
		// Catat percobaan gagal
		attempts++
		// Gunakan Set karena SetTTL tidak tersedia di interface Storage
		err = c.store.Set(bgCtx, ipKey, []byte(strconv.Itoa(attempts)))
		if err != nil {
			c.logger.WithError(err).Warn("Gagal menyimpan percobaan login")
		}

		c.logger.Warn("Percobaan login gagal", utils.Fields{
			"ip":       ctx.IP(),
			"attempts": attempts,
		})
		return ctx.Redirect("/login?error=invalid_credentials&redirect=" + redirect)
	}

	// Token valid, reset percobaan
	c.store.Delete(bgCtx, ipKey)

	// Buat sesi
	sess, err = c.sessionStore.Get(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mendapatkan sesi saat login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan sesi")
	}

	// Set token dan status autentikasi di sesi
	sess.Set("auth_token", token)
	sess.Set("authenticated", true)
	sess.Set("last_activity_time", time.Now().Unix())

	// Simpan fingerprint perangkat
	deviceFingerprint := fmt.Sprintf("%s|%s", ctx.IP(), ctx.Get("User-Agent"))
	sess.Set("device_fingerprint", deviceFingerprint)

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

	// Tambahkan script untuk set token di localStorage untuk API calls
	redirectHTML := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>Redirecting...</title></head>
	<body>
		<script>
			// Set token for API calls
			localStorage.setItem('access_token', '%s');
			sessionStorage.setItem('authenticated', 'true');
			// Redirect to requested page
			window.location.href = '%s';
		</script>
	</body>
	</html>
	`, token, redirect)

	ctx.Set("Content-Type", "text/html")
	return ctx.SendString(redirectHTML)
}

// Logout menangani logout pengguna
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Dapatkan dan hapus sesi
	sess, err := c.sessionStore.Get(ctx)
	if err == nil {
		sess.Delete("auth_token")
		sess.Delete("authenticated")
		sess.Delete("last_activity_time")
		sess.Delete("device_fingerprint")
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

// Fungsi helper untuk menghasilkan random token untuk CSRF
func generateRandomToken(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// maskToken menyembunyikan sebagian besar token dengan * untuk keamanan
func (c *AuthController) maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}

	// Tampilkan hanya 4 karakter pertama dan 4 terakhir
	return token[:4] + "..." + token[len(token)-4:]
}
