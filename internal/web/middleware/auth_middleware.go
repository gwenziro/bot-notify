package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/auth"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthMiddleware menangani autentikasi untuk route web
type AuthMiddleware struct {
	config       *config.Config
	sessionStore *session.Store
	storage      storage.Storage
	validator    *auth.AuthValidator
	tokenManager *auth.TokenManager
	sessionMgr   *auth.SessionManager
	logger       utils.LogrusEntry
}

// NewAuthMiddleware membuat instance baru AuthMiddleware
func NewAuthMiddleware(cfg *config.Config, sessionStore *session.Store, storage storage.Storage) *AuthMiddleware {
	logger := utils.ForModule("auth-middleware")

	// Buat TokenManager
	tokenManager := auth.NewTokenManager(&cfg.Auth, logger)

	// Buat SessionManager
	sessionManager := auth.NewSessionManager(sessionStore, storage, &cfg.Auth, logger)

	// Buat AuthValidator
	validator := auth.NewAuthValidator(&cfg.Auth, tokenManager, sessionManager, logger)

	return &AuthMiddleware{
		config:       cfg,
		sessionStore: sessionStore,
		storage:      storage,
		validator:    validator,
		tokenManager: tokenManager,
		sessionMgr:   sessionManager,
		logger:       logger,
	}
}

// RequireAuth middleware untuk mengharuskan autentikasi
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Validasi akses web
		err := m.validator.ValidateWebAccess(c)
		if err != nil {
			// Log percobaan akses tanpa autentikasi
			m.logger.WithFields(utils.Fields{
				"path":    c.Path(),
				"ip":      c.IP(),
				"referer": c.Get("Referer"),
			}).Info("Akses ditolak: autentikasi diperlukan")

			// Redirect ke halaman login dengan URL callback
			redirectURL := "/login?redirect=" + c.Path()
			return c.Redirect(redirectURL, fiber.StatusFound)
		}

		// Jika autentikasi berhasil, perpanjang sesi jika mendekati kedaluwarsa
		if m.sessionShouldBeRenewed(c) {
			m.renewSession(c)
		}

		// Atur user info di locals
		userID, _ := m.sessionMgr.GetUserID(c)
		c.Locals("user", fiber.Map{
			"id":        userID,
			"name":      "Admin", // Dapat diubah jika multi-user diterapkan
			"logged_in": true,
		})

		// Lanjutkan ke handler berikutnya
		return c.Next()
	}
}

// ValidateAPIToken middleware untuk validasi token API
func (m *AuthMiddleware) ValidateAPIToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Catat waktu mulai untuk logging
		startTime := time.Now()

		// Validasi token API
		err := m.validator.ValidateAPIAccess(c)
		if err != nil {
			// Log percobaan akses API yang gagal
			m.logger.WithFields(utils.Fields{
				"path":    c.Path(),
				"ip":      utils.MaskIP(c.IP()),
				"method":  c.Method(),
				"error":   err.Error(),
				"latency": time.Since(startTime).String(),
			}).Warn("Akses API ditolak: token tidak valid")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token akses tidak valid",
				"code":    "INVALID_TOKEN",
			})
		}

		// Log successful API request
		m.logger.WithFields(utils.Fields{
			"path":    c.Path(),
			"method":  c.Method(),
			"latency": time.Since(startTime).String(),
		}).Debug("API request berhasil divalidasi")

		// Lanjutkan ke handler berikutnya
		return c.Next()
	}
}

// OptionalAuth middleware untuk mengecek autentikasi tanpa mengharuskannya
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Coba validasi akses web
		err := m.validator.ValidateWebAccess(c)
		if err == nil {
			// Atur user info di locals jika terautentikasi
			userID, _ := m.sessionMgr.GetUserID(c)
			c.Locals("user", fiber.Map{
				"id":        userID,
				"name":      "Admin",
				"logged_in": true,
			})
		} else {
			c.Locals("user", fiber.Map{
				"logged_in": false,
			})
		}

		// Lanjutkan ke handler berikutnya
		return c.Next()
	}
}

// GetAuthValidator mengembalikan validator autentikasi
func (m *AuthMiddleware) GetAuthValidator() *auth.AuthValidator {
	return m.validator
}

// GetTokenManager mengembalikan token manager
func (m *AuthMiddleware) GetTokenManager() *auth.TokenManager {
	return m.tokenManager
}

// GetSessionManager mengembalikan session manager
func (m *AuthMiddleware) GetSessionManager() *auth.SessionManager {
	return m.sessionMgr
}

// sessionShouldBeRenewed memeriksa apakah sesi perlu diperpanjang
func (m *AuthMiddleware) sessionShouldBeRenewed(c *fiber.Ctx) bool {
	sess, err := m.sessionStore.Get(c)
	if err != nil {
		return false
	}

	// Dapatkan waktu kedaluwarsa sesi
	expiresAt := sess.Get("expires_at")
	if expiresAt == nil {
		return true
	}

	expiry, ok := expiresAt.(int64)
	if !ok {
		return true
	}

	// Jika waktu kedaluwarsa kurang dari 25% dari waktu total, perpanjang
	expiryTime := time.Unix(expiry, 0)
	sessionDuration := m.config.Auth.TokenExpiry
	if time.Until(expiryTime) < sessionDuration/4 {
		return true
	}

	return false
}

// renewSession memperpanjang sesi yang akan kedaluwarsa
func (m *AuthMiddleware) renewSession(c *fiber.Ctx) {
	sess, err := m.sessionStore.Get(c)
	if err != nil {
		return
	}

	// Perbarui waktu kedaluwarsa
	newExpiry := time.Now().Add(m.config.Auth.TokenExpiry)
	sess.Set("expires_at", newExpiry.Unix())

	// Simpan sesi
	if err := sess.Save(); err != nil {
		m.logger.WithError(err).Error("Gagal menyimpan sesi yang diperpanjang")
	}

	// Perbarui juga di storage
	m.sessionMgr.UpdateSessionActivity(c)
}
