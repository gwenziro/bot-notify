package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthMiddleware mengelola autentikasi untuk web UI
type AuthMiddleware struct {
	config       *config.Config
	sessionStore *session.Store
	logger       utils.LogrusEntry
}

// NewAuthMiddleware membuat instance baru AuthMiddleware
func NewAuthMiddleware(cfg *config.Config, sessionStore *session.Store) *AuthMiddleware {
	return &AuthMiddleware{
		config:       cfg,
		sessionStore: sessionStore,
		logger:       utils.ForModule("auth-middleware"),
	}
}

// RequireAuth memvalidasi bahwa pengguna telah login
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Debug log untuk melihat request yang masuk
		m.logger.Debug("Checking authentication", utils.Fields{
			"path":   c.Path(),
			"method": c.Method(),
			"ip":     c.IP(),
		})

		// Dapatkan sesi dengan error handling yang lebih baik
		sess, err := m.sessionStore.Get(c)
		if err != nil {
			m.logger.Error("Failed to get session", utils.Fields{
				"error": err.Error(),
				"path":  c.Path(),
				"ip":    c.IP(),
			})
			return m.redirectToLogin(c, "invalid_session")
		}

		// Periksa apakah pengguna telah login
		authenticated := sess.Get("authenticated")
		authToken := sess.Get("auth_token")

		m.logger.Debug("Session data", utils.Fields{
			"authenticated": authenticated,
			"has_token":     authToken != nil,
			"session_id":    sess.ID(),
		})

		// Jika tidak ada status authenticated, coba auto-login dari cookie
		if authenticated == nil || authenticated == false {
			m.logger.Debug("No authentication found in session, checking auto-login cookie")

			autoLoginToken := c.Cookies("auto_login")
			if autoLoginToken != "" {
				m.logger.Debug("Found auto-login cookie, attempting auto-login")

				if m.secureCompare(autoLoginToken, m.config.Auth.AccessToken) {
					// Auto-login berhasil, set session
					sess.Set("auth_token", m.config.Auth.AccessToken)
					sess.Set("authenticated", true)
					sess.Set("last_activity_time", time.Now().Unix())

					// Simpan fingerprint perangkat
					deviceFingerprint := fmt.Sprintf("%s|%s", c.IP(), c.Get("User-Agent"))
					sess.Set("device_fingerprint", deviceFingerprint)

					if err := sess.Save(); err != nil {
						m.logger.Error("Failed to save auto-login session", utils.Fields{
							"error": err.Error(),
						})
						return m.redirectToLogin(c, "session_save_error")
					}

					m.logger.Info("Auto-login successful", utils.Fields{
						"path": c.Path(),
						"ip":   c.IP(),
					})

					// Set security headers dan lanjutkan
					m.addSecurityHeaders(c)
					return c.Next()
				} else {
					m.logger.Warn("Invalid auto-login token", utils.Fields{
						"ip":   c.IP(),
						"path": c.Path(),
					})
				}
			}

			// Tidak ada auto-login yang valid, redirect ke login
			return m.redirectToLogin(c, "not_authenticated")
		}

		// Verifikasi token jika ada
		if authToken == nil {
			m.logger.Warn("No auth token in session", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})
			return m.redirectToLogin(c, "no_token")
		}

		// Verifikasi token dengan secure compare
		token, ok := authToken.(string)
		if !ok || !m.secureCompare(token, m.config.Auth.AccessToken) {
			m.logger.Warn("Invalid token", utils.Fields{
				"path":       c.Path(),
				"ip":         c.IP(),
				"token_type": fmt.Sprintf("%T", authToken),
			})

			// Clear invalid session
			sess.Delete("auth_token")
			sess.Delete("authenticated")
			sess.Delete("last_activity_time")
			sess.Delete("device_fingerprint")
			sess.Save()

			return m.redirectToLogin(c, "invalid_token")
		}

		// Periksa session timeout
		if err := m.checkSessionTimeout(sess); err != nil {
			m.logger.Info("Session timeout", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})
			return m.redirectToLogin(c, "session_timeout")
		}

		// Periksa device fingerprint untuk security
		if err := m.checkDeviceFingerprint(c, sess); err != nil {
			m.logger.Warn("Device fingerprint mismatch", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})
			return m.redirectToLogin(c, "security_concern")
		}

		// Update last activity
		sess.Set("last_activity_time", time.Now().Unix())
		if err := sess.Save(); err != nil {
			m.logger.Error("Failed to update session activity", utils.Fields{
				"error": err.Error(),
			})
		}

		// Set security headers
		m.addSecurityHeaders(c)

		m.logger.Debug("Authentication successful", utils.Fields{
			"path": c.Path(),
			"ip":   c.IP(),
		})

		return c.Next()
	}
}

// secureCompare membandingkan dua string dengan waktu konstan
// untuk mencegah timing attack
func (m *AuthMiddleware) secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// redirectToLogin mengarahkan ke halaman login dengan parameter error
func (m *AuthMiddleware) redirectToLogin(c *fiber.Ctx, errorType string) error {
	// Untuk AJAX request, return JSON error
	if c.Get("X-Requested-With") == "XMLHttpRequest" ||
		c.Get("Accept") == "application/json" ||
		c.Get("Content-Type") == "application/json" {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
			"code":    401,
		})
	}

	// Untuk request biasa, redirect ke login
	redirectURL := fmt.Sprintf("/login?redirect=%s&error=%s", c.Path(), errorType)
	return c.Redirect(redirectURL)
}

// checkSessionTimeout memeriksa apakah session sudah timeout
func (m *AuthMiddleware) checkSessionTimeout(sess *session.Session) error {
	lastActivityTime := sess.Get("last_activity_time")
	if lastActivityTime == nil {
		return fmt.Errorf("no last activity time")
	}

	lastActivity, ok := lastActivityTime.(int64)
	if !ok {
		return fmt.Errorf("invalid last activity time format")
	}

	// Session timeout 30 menit
	maxInactiveInterval := int64(30 * 60) // 30 menit dalam detik
	if time.Now().Unix()-lastActivity > maxInactiveInterval {
		// Clear session
		sess.Delete("auth_token")
		sess.Delete("authenticated")
		sess.Delete("last_activity_time")
		sess.Delete("device_fingerprint")
		sess.Save()
		return fmt.Errorf("session timeout")
	}

	return nil
}

// checkDeviceFingerprint memeriksa konsistensi device fingerprint
func (m *AuthMiddleware) checkDeviceFingerprint(c *fiber.Ctx, sess *session.Session) error {
	currentFingerprint := fmt.Sprintf("%s|%s", c.IP(), c.Get("User-Agent"))
	storedFingerprint := sess.Get("device_fingerprint")

	if storedFingerprint == nil {
		// Set fingerprint jika belum ada
		sess.Set("device_fingerprint", currentFingerprint)
		return nil
	}

	if storedFingerprint.(string) != currentFingerprint {
		// Clear session karena kemungkinan session hijacking
		sess.Delete("auth_token")
		sess.Delete("authenticated")
		sess.Delete("last_activity_time")
		sess.Delete("device_fingerprint")
		sess.Save()
		return fmt.Errorf("device fingerprint mismatch")
	}

	return nil
}

// addSecurityHeaders menambahkan header keamanan
func (m *AuthMiddleware) addSecurityHeaders(c *fiber.Ctx) {
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: img.icons8.com")
}

// SetAutoLogin menetapkan cookie auto-login
func (m *AuthMiddleware) SetAutoLogin(c *fiber.Ctx, enabled bool) {
	if enabled {
		c.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    m.config.Auth.AccessToken,
			Path:     "/",
			MaxAge:   m.config.Auth.CookieMaxAge,
			Secure:   m.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	} else {
		c.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   m.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	}
}
