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

// RequireAuth memvalidasi bahwa pengguna telah login
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Dapatkan sesi
		sess, err := m.sessionStore.Get(c)
		if err != nil {
			m.logger.WithError(err).Warn("Gagal mendapatkan sesi")
			return c.Redirect("/login?redirect=" + c.Path())
		}

		// Periksa apakah pengguna telah login
		authToken := sess.Get("auth_token")
		if authToken == nil {
			// Coba dapatkan auto-login token dari cookie
			autoLoginToken := c.Cookies("auto_login")
			if autoLoginToken != "" && m.secureCompare(autoLoginToken, m.config.Auth.AccessToken) {
				// Auto-login berhasil
				sess.Set("auth_token", m.config.Auth.AccessToken)
				sess.Set("authenticated", true)
				sess.Set("last_activity_time", time.Now().Unix())

				// Simpan fingerprint perangkat
				deviceFingerprint := fmt.Sprintf("%s|%s", c.IP(), c.Get("User-Agent"))
				sess.Set("device_fingerprint", deviceFingerprint)

				if err := sess.Save(); err != nil {
					m.logger.WithError(err).Warn("Gagal menyimpan sesi auto-login")
				}

				m.logger.Info("Auto-login berhasil", utils.Fields{
					"path": c.Path(),
					"ip":   c.IP(),
				})

				// Tambahkan security headers
				addSecurityHeaders(c)

				return c.Next()
			}

			// Tidak ada token, redirect ke login
			return c.Redirect("/login?redirect=" + c.Path())
		}

		// Verifikasi token dengan waktu konstan untuk mencegah timing attack
		token := authToken.(string)
		if !m.secureCompare(token, m.config.Auth.AccessToken) {
			sess.Delete("auth_token")
			sess.Delete("authenticated")
			sess.Delete("last_activity_time")
			sess.Delete("device_fingerprint")
			if err := sess.Save(); err != nil {
				m.logger.WithError(err).Warn("Gagal menghapus sesi invalid")
			}
			return c.Redirect("/login?error=invalid_session&redirect=" + c.Path())
		}

		// Tambahkan pemeriksaan session timeout
		lastActivityTime := sess.Get("last_activity_time")
		if lastActivityTime != nil {
			lastActivity, ok := lastActivityTime.(int64)
			if ok {
				// Paksa logout jika tidak ada aktivitas selama 30 menit
				maxInactiveInterval := int64(30 * 60) // 30 menit dalam detik
				if time.Now().Unix()-lastActivity > maxInactiveInterval {
					m.logger.Info("Sesi kedaluwarsa karena tidak aktif")
					sess.Delete("auth_token")
					sess.Delete("authenticated")
					sess.Delete("last_activity_time")
					sess.Delete("device_fingerprint")
					sess.Save()
					return c.Redirect("/login?error=session_timeout&redirect=" + c.Path())
				}
			}
		}

		// Perbarui waktu aktivitas terakhir
		sess.Set("last_activity_time", time.Now().Unix())
		sess.Save()

		// Periksa fingerprint perangkat untuk deteksi session hijacking
		currentFingerprint := fmt.Sprintf("%s|%s", c.IP(), c.Get("User-Agent"))
		storedFingerprint := sess.Get("device_fingerprint")

		if storedFingerprint == nil {
			// Fingerprint belum ada, simpan yang baru
			sess.Set("device_fingerprint", currentFingerprint)
			sess.Save()
		} else if storedFingerprint.(string) != currentFingerprint {
			// Fingerprint berbeda, kemungkinan session hijacking
			m.logger.Warn("Kemungkinan session hijacking terdeteksi", utils.Fields{
				"stored_fingerprint":  storedFingerprint.(string),
				"current_fingerprint": currentFingerprint,
			})

			// Hapus sesi dan arahkan ke login
			sess.Delete("auth_token")
			sess.Delete("authenticated")
			sess.Delete("last_activity_time")
			sess.Delete("device_fingerprint")
			sess.Save()

			return c.Redirect("/login?error=security_concern&redirect=" + c.Path())
		}

		// Tambahkan security headers
		addSecurityHeaders(c)

		return c.Next()
	}
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

// Helper function untuk menambahkan security headers
func addSecurityHeaders(c *fiber.Ctx) {
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: img.icons8.com")
}
