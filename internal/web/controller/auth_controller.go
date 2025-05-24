package controller

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/auth"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthController menangani operasi autentikasi
type AuthController struct {
	config         *config.Config
	whatsapp       *client.Client
	sessionStore   *session.Store
	storage        storage.Storage
	tokenManager   *auth.TokenManager
	sessionManager *auth.SessionManager
	validator      *auth.AuthValidator
	logger         utils.LogrusEntry
	// Tambahkan ratelimiter dan failed login tracker
	loginRateLimiter *utils.RateLimiter
}

// NewAuthController membuat instance baru AuthController
func NewAuthController(cfg *config.Config, whatsapp *client.Client, sessionStore *session.Store, storage storage.Storage, logger utils.LogrusEntry) *AuthController {
	authLogger := logger.WithField("component", "auth-controller")

	// Buat token manager
	tokenManager := auth.NewTokenManager(&cfg.Auth, authLogger)

	// Buat session manager
	sessionManager := auth.NewSessionManager(sessionStore, storage, &cfg.Auth, authLogger)

	// Buat validator
	validator := auth.NewAuthValidator(&cfg.Auth, tokenManager, sessionManager, authLogger)

	// Buat ratelimiter untuk mencegah brute force
	loginRateLimiter := utils.NewRateLimiter(5, time.Minute) // 5 request per menit per IP

	return &AuthController{
		config:           cfg,
		whatsapp:         whatsapp,
		sessionStore:     sessionStore,
		storage:          storage,
		tokenManager:     tokenManager,
		sessionManager:   sessionManager,
		validator:        validator,
		logger:           authLogger,
		loginRateLimiter: loginRateLimiter,
	}
}

// LoginPage menampilkan halaman login
func (c *AuthController) LoginPage(ctx *fiber.Ctx) error {
	// Periksa apakah pengguna sudah terautentikasi
	if c.sessionManager.IsAuthenticated(ctx) {
		// Jika sudah login, redirect ke dashboard
		return ctx.Redirect("/dashboard", fiber.StatusFound)
	}

	// Generate CSRF token
	csrfToken, err := utils.GenerateCSRFToken()
	if err != nil {
		c.logger.WithError(err).Error("Gagal generate CSRF token")
		csrfToken = fmt.Sprintf("%d", time.Now().UnixNano()) // Fallback sederhana
	}

	// Siapkan data untuk view
	return ctx.Render("auth/login", fiber.Map{
		"Title":      "Login",
		"RedirectTo": ctx.Query("redirect", "/dashboard"),
		"Error":      ctx.Query("error"),
		"CSRFToken":  csrfToken,
	})
}

// ProcessLogin memproses form login
func (c *AuthController) ProcessLogin(ctx *fiber.Ctx) error {
	// Get client IP untuk rate limiting dan logging
	clientIP := ctx.IP()

	// Cek rate limit untuk mencegah brute force
	if c.loginRateLimiter.ExceedsLimit(clientIP) {
		c.logger.WithFields(utils.Fields{
			"ip": utils.MaskIP(clientIP),
		}).Warn("Login rate limit exceeded")

		return ctx.Redirect("/login?error=Terlalu banyak percobaan login. Silakan coba lagi setelah beberapa saat.")
	}

	// Validasi CSRF token
	submittedToken := ctx.FormValue("csrf_token")
	expectedToken := ctx.Locals("csrf_token")

	if submittedToken == "" || expectedToken == nil || submittedToken != expectedToken.(string) {
		c.logger.WithFields(utils.Fields{
			"ip": utils.MaskIP(clientIP),
		}).Warn("Invalid CSRF token in login attempt")

		return ctx.Redirect("/login?error=Permintaan tidak valid. Silakan coba lagi.")
	}

	// Ambil token dari form
	token := ctx.FormValue("token")
	rememberMe := ctx.FormValue("remember_me") == "on"
	redirectTo := ctx.FormValue("redirect", "/dashboard")

	// Track login attempt
	c.loginRateLimiter.AddRequest(clientIP)

	// Validasi token
	err := c.validator.ValidateLogin(token)
	if err != nil {
		// Cek untuk brute force protection
		isBlocked, duration := c.validator.CheckPasswordBruteForce(ctx)
		if isBlocked {
			c.logger.WithFields(utils.Fields{
				"ip":       utils.MaskIP(clientIP),
				"duration": duration.Error(),
			}).Warn("IP diblokir karena terlalu banyak percobaan login gagal")

			return ctx.Redirect("/login?error=Terlalu banyak percobaan login. Silakan coba lagi setelah beberapa saat.")
		}

		// Log percobaan login gagal
		c.logger.WithFields(utils.Fields{
			"ip":       utils.MaskIP(clientIP),
			"redirect": redirectTo,
			"ua":       ctx.Get("User-Agent"),
		}).Warn("Percobaan login gagal: token tidak valid")

		return ctx.Redirect("/login?error=Token akses tidak valid&redirect=" + redirectTo)
	}

	// Buat sesi baru
	sess, err := c.sessionManager.CreateSession(ctx, "admin", rememberMe)
	if err != nil {
		c.logger.WithError(err).Error("Gagal membuat sesi")
		return ctx.Redirect("/login?error=Gagal membuat sesi, silakan coba lagi")
	}

	// Simpan token sesi di cookie dengan opsi keamanan
	cookie := fiber.Cookie{
		Name:     c.config.Auth.CookieName,
		Value:    sess.ID(),
		Expires:  time.Now().Add(c.config.Auth.TokenExpiry),
		HTTPOnly: true,
		Secure:   c.config.Auth.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
	}

	// Jika remember me, set expiry yang lebih lama
	if rememberMe {
		cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	}

	ctx.Cookie(&cookie)

	// Log login berhasil
	c.logger.WithFields(utils.Fields{
		"ip":         utils.MaskIP(clientIP),
		"user_agent": ctx.Get("User-Agent"),
		"redirect":   redirectTo,
		"remember":   rememberMe,
	}).Info("Login berhasil")

	// Redirect ke halaman yang diminta
	return ctx.Redirect(redirectTo)
}

// Logout melakukan logout pengguna
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Log logout request
	c.logger.WithFields(utils.Fields{
		"ip": utils.MaskIP(ctx.IP()),
		"ua": ctx.Get("User-Agent"),
	}).Info("Logout request diterima")

	// Dapatkan sesi aktif
	sess, err := c.sessionStore.Get(ctx)
	if err == nil {
		// Log info sesi
		userID := sess.Get("user_id")
		c.logger.WithFields(utils.Fields{
			"user_id": userID,
			"ip":      utils.MaskIP(ctx.IP()),
		}).Info("Menghapus sesi pengguna")
	}

	// Hapus sesi
	err = c.sessionManager.DestroySession(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Gagal menghapus sesi")
	}

	// Hapus cookie
	ctx.Cookie(&fiber.Cookie{
		Name:     c.config.Auth.CookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expired di masa lalu
		HTTPOnly: true,
		Secure:   c.config.Auth.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
	})

	// Redirect ke halaman login
	return ctx.Redirect("/login")
}

// ValidateToken endpoint untuk memvalidasi token secara programatis
func (c *AuthController) ValidateToken(ctx *fiber.Ctx) error {
	// Validasi token API
	err := c.validator.ValidateAPIAccess(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"valid": false,
			"error": err.Error(),
			"code":  "INVALID_TOKEN",
		})
	}

	return ctx.JSON(fiber.Map{
		"valid": true,
	})
}

// RefreshSession memperpanjang sesi yang akan kedaluwarsa
func (c *AuthController) RefreshSession(ctx *fiber.Ctx) error {
	// Periksa apakah pengguna terautentikasi
	if !c.sessionManager.IsAuthenticated(ctx) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Tidak terautentikasi",
			"code":    "NOT_AUTHENTICATED",
		})
	}

	// Ambil sesi yang ada
	sess, err := c.sessionStore.Get(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mendapatkan sesi",
			"code":    "SESSION_ERROR",
		})
	}

	// Dapatkan waktu kedaluwarsa saat ini
	expiresAt := sess.Get("expires_at")
	if expiresAt == nil {
		// Jika tidak ada, atur expiry baru
		sess.Set("expires_at", time.Now().Add(c.config.Auth.TokenExpiry).Unix())
	} else {
		// Jika ada, periksa apakah akan kedaluwarsa dalam 2 jam
		expiry, ok := expiresAt.(int64)
		if ok && time.Now().Add(2*time.Hour).Unix() > expiry {
			// Perbaharui waktu kedaluwarsa
			sess.Set("expires_at", time.Now().Add(c.config.Auth.TokenExpiry).Unix())
		} else if ok {
			// Sesi masih aktif dan jauh dari kedaluwarsa
			return ctx.JSON(fiber.Map{
				"success": true,
				"message": "Sesi masih aktif",
				"expires": time.Unix(expiry, 0),
			})
		}
	}

	// Simpan sesi
	if err := sess.Save(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan sesi",
			"code":    "SESSION_SAVE_ERROR",
		})
	}

	// Perbarui juga data sesi di storage
	if err := c.sessionManager.UpdateSessionActivity(ctx); err != nil {
		c.logger.WithError(err).Error("Gagal memperbarui data sesi")
	}

	// Perbarui cookie sesi
	cookie := fiber.Cookie{
		Name:     c.config.Auth.CookieName,
		Value:    sess.ID(),
		Expires:  time.Now().Add(c.config.Auth.TokenExpiry),
		HTTPOnly: true,
		Secure:   c.config.Auth.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
	}
	ctx.Cookie(&cookie)

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Sesi diperpanjang",
		"expires": time.Unix(sess.Get("expires_at").(int64), 0),
	})
}

// GenerateNewAPIToken membuat token API baru
func (c *AuthController) GenerateNewAPIToken(ctx *fiber.Ctx) error {
	// Validasi akses admin
	if !c.sessionManager.IsAuthenticated(ctx) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Tidak terautentikasi",
		})
	}

	// Generate token baru
	token, err := c.validator.GenerateSecureToken()
	if err != nil {
		c.logger.WithError(err).Error("Gagal generate token API baru")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal generate token",
		})
	}

	// Hash token untuk penyimpanan
	hashedToken, err := c.tokenManager.SetAPIToken(token)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal hash token",
		})
	}

	// Simpan ke config (implementasi tergantung pada config store)
	c.config.Auth.HashedAccessToken = hashedToken
	c.config.Auth.AccessToken = token // Simpan token plaintext di memory untuk validasi

	// TODO: Simpan config ke file

	return ctx.JSON(fiber.Map{
		"success": true,
		"token":   token,
		"message": "Token API baru berhasil dibuat",
	})
}

// HandleSessionError menangani kesalahan sesi
func (c *AuthController) HandleSessionError(ctx *fiber.Ctx, err error) error {
	c.logger.WithError(err).Error("Kesalahan sesi")

	// Hapus cookie sesi yang bermasalah
	ctx.Cookie(&fiber.Cookie{
		Name:     c.config.Auth.CookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Kedaluwarsa di masa lalu
		HTTPOnly: true,
		Path:     "/",
	})

	// Redirect ke halaman login
	return ctx.Redirect("/login?error=Sesi tidak valid, silakan login kembali")
}
