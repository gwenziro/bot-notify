package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
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

	// Dapatkan data untuk halaman login
	pageData := c.prepareLoginPageData(ctx)

	// Tambahkan log untuk debugging
	c.logger.Info("Rendering halaman login", utils.Fields{
		"redirect": pageData.RedirectTo,
		"error":    pageData.Error,
	})

	// Render halaman login
	return ctx.Render("login", pageData)
}

// prepareLoginPageData menyiapkan data untuk halaman login
func (c *AuthController) prepareLoginPageData(ctx *fiber.Ctx) entity.LoginPageData {
	// Dapatkan parameter
	redirect := ctx.Query("redirect", "/dashboard")
	errorMsg := ctx.Query("error", "")

	// Tentukan pesan error berdasarkan kode
	errorText := c.getErrorMessage(errorMsg)

	// Generate CSRF token dan simpan dalam sesi
	csrfToken := utils.GenerateRandomToken(32)
	sess, err := c.sessionStore.Get(ctx)
	if err == nil {
		sess.Set("csrf_token", csrfToken)
		if err := sess.Save(); err != nil {
			c.logger.WithError(err).Warn("Gagal menyimpan CSRF token ke sesi")
		}
	}

	// Buat data halaman login
	return entity.LoginPageData{
		Title:       "Login - WhatsApp Bot Notify",
		RedirectTo:  redirect,
		Error:       errorText,
		CsrfToken:   csrfToken,
		AccessToken: utils.MaskToken(c.config.Auth.AccessToken),
	}
}

// getErrorMessage menerjemahkan kode error menjadi pesan yang bisa dibaca
func (c *AuthController) getErrorMessage(errorCode string) string {
	switch errorCode {
	case "invalid_credentials":
		return "Token akses tidak valid. Silakan coba lagi."
	case "invalid_session":
		return "Sesi Anda telah kedaluwarsa. Silakan login kembali."
	case "too_many_attempts":
		return "Terlalu banyak percobaan login gagal. Silakan coba lagi nanti."
	case "session_timeout":
		return "Sesi Anda telah kedaluwarsa karena tidak aktif terlalu lama."
	case "security_concern":
		return "Terdapat masalah keamanan dengan sesi Anda. Silakan login kembali."
	case "invalid_request":
		return "Permintaan tidak valid. Silakan coba lagi."
	default:
		return ""
	}
}

// ProcessLogin memproses login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Verifikasi CSRF token
	if !c.verifyCsrfToken(ctx) {
		return ctx.Redirect("/login?error=invalid_request&redirect=" + ctx.FormValue("redirect", "/dashboard"))
	}

	// Validasi rate limiting
	if !c.validateRateLimit(ctx) {
		return ctx.Redirect("/login?error=too_many_attempts&redirect=" + ctx.FormValue("redirect", "/dashboard"))
	}

	// Dapatkan data login dari form
	loginReq := entity.LoginRequest{
		Token:      ctx.FormValue("token"),
		RememberMe: ctx.FormValue("remember_me") == "on",
		Redirect:   ctx.FormValue("redirect", "/dashboard"),
	}

	// Validasi token
	if loginReq.Token != c.config.Auth.AccessToken {
		c.incrementLoginAttempt(ctx)
		c.logger.Warn("Percobaan login gagal", utils.Fields{
			"ip": ctx.IP(),
		})
		return ctx.Redirect("/login?error=invalid_credentials&redirect=" + loginReq.Redirect)
	}

	// Token valid, reset percobaan
	c.resetLoginAttempt(ctx)

	// Buat sesi
	if err := c.createSession(ctx, loginReq.Token); err != nil {
		c.logger.WithError(err).Error("Gagal membuat sesi login")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Terjadi kesalahan saat menyimpan sesi")
	}

	// Set cookie auto-login jika "remember me" dicentang
	if loginReq.RememberMe {
		c.setAutoLoginCookie(ctx)
	}

	c.logger.Info("Login berhasil", utils.Fields{
		"ip":          ctx.IP(),
		"remember_me": loginReq.RememberMe,
	})

	// Redirect dengan script untuk set token di localStorage
	return c.renderRedirectPage(ctx, loginReq.Token, loginReq.Redirect)
}

// verifyCsrfToken memeriksa validitas token CSRF
func (c *AuthController) verifyCsrfToken(ctx *fiber.Ctx) bool {
	csrfToken := ctx.FormValue("csrf_token")
	sess, err := c.sessionStore.Get(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mendapatkan sesi saat validasi CSRF")
		return false
	}

	storedToken := sess.Get("csrf_token")
	if storedToken == nil || csrfToken == "" || csrfToken != storedToken {
		c.logger.Warn("CSRF token tidak valid", utils.Fields{
			"ip": ctx.IP(),
		})
		return false
	}

	// Hapus CSRF token setelah digunakan
	sess.Delete("csrf_token")
	sess.Save()

	return true
}

// validateRateLimit memeriksa apakah pengguna telah melebihi batas percobaan login
func (c *AuthController) validateRateLimit(ctx *fiber.Ctx) bool {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()

	attemptData, err := c.store.Get(bgCtx, ipKey)
	if err != nil {
		return true // Asumsikan valid jika gagal mendapatkan data
	}

	if len(attemptData) > 0 {
		attempts, _ := strconv.Atoi(string(attemptData))
		if attempts >= 5 {
			c.logger.Warn("Login rate limited", utils.Fields{
				"ip":       ctx.IP(),
				"attempts": attempts,
			})
			return false
		}
	}

	return true
}

// incrementLoginAttempt menambahkan hitungan percobaan login gagal
func (c *AuthController) incrementLoginAttempt(ctx *fiber.Ctx) {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()

	attemptData, err := c.store.Get(bgCtx, ipKey)
	attempts := 1
	if err == nil && len(attemptData) > 0 {
		attempts, _ = strconv.Atoi(string(attemptData))
		attempts++
	}

	err = c.store.Set(bgCtx, ipKey, []byte(strconv.Itoa(attempts)))
	if err != nil {
		c.logger.WithError(err).Warn("Gagal menyimpan percobaan login")
	}
}

// resetLoginAttempt menghapus hitungan percobaan login
func (c *AuthController) resetLoginAttempt(ctx *fiber.Ctx) {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()
	c.store.Delete(bgCtx, ipKey)
}

// createSession membuat sesi login baru
func (c *AuthController) createSession(ctx *fiber.Ctx, token string) error {
	sess, err := c.sessionStore.Get(ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	// Set token dan status autentikasi di sesi
	sess.Set("auth_token", token)
	sess.Set("authenticated", true)
	sess.Set("last_activity_time", time.Now().Unix())

	// Simpan fingerprint perangkat
	deviceFingerprint := fmt.Sprintf("%s|%s", ctx.IP(), ctx.Get("User-Agent"))
	sess.Set("device_fingerprint", deviceFingerprint)

	return sess.Save()
}

// setAutoLoginCookie menetapkan cookie untuk auto-login
func (c *AuthController) setAutoLoginCookie(ctx *fiber.Ctx) {
	cookie := entity.AutoLoginCookie{
		Name:     "auto_login",
		Value:    c.config.Auth.AccessToken,
		Path:     "/",
		MaxAge:   c.config.Auth.CookieMaxAge,
		Secure:   c.config.Server.BaseURL != "http://localhost:8080",
		HTTPOnly: true,
		SameSite: "Strict",
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		MaxAge:   cookie.MaxAge,
		Secure:   cookie.Secure,
		HTTPOnly: cookie.HTTPOnly,
		SameSite: cookie.SameSite,
	})
}

// renderRedirectPage menampilkan halaman redirect dengan script untuk set token
func (c *AuthController) renderRedirectPage(ctx *fiber.Ctx, token string, redirect string) error {
	redirectHTML := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Redirecting...</title>
		<script>
			// Fungsi untuk memastikan token disimpan dengan benar
			function ensureTokenSaved(token) {
				localStorage.setItem('access_token', token);
				
				// Verifikasi token tersimpan dengan benar
				const savedToken = localStorage.getItem('access_token');
				if (savedToken !== token) {
					console.error('Token tidak tersimpan dengan benar di localStorage');
					alert('Terjadi masalah saat menyimpan token autentikasi. Mohon reload halaman jika mengalami masalah.');
				} else {
					console.log('Token berhasil disimpan di localStorage');
				}
				
				sessionStorage.setItem('authenticated', 'true');
			}
			
			// Simpan token untuk API calls
			ensureTokenSaved('%s');
			
			// Redirect ke halaman yang diminta
			window.location.href = '%s';
		</script>
	</head>
	<body>
		<p>Redirecting to dashboard...</p>
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
