package middleware

import (
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
			if autoLoginToken != "" && autoLoginToken == m.config.Auth.AccessToken {
				// Auto-login berhasil
				sess.Set("auth_token", m.config.Auth.AccessToken)
				sess.Set("authenticated", true)
				if err := sess.Save(); err != nil {
					m.logger.WithError(err).Warn("Gagal menyimpan sesi auto-login")
				}

				m.logger.Info("Auto-login berhasil", utils.Fields{
					"path": c.Path(),
					"ip":   c.IP(),
				})
				return c.Next()
			}

			// Tidak ada token, redirect ke login
			return c.Redirect("/login?redirect=" + c.Path())
		}

		// Verifikasi token
		token := authToken.(string)
		if token != m.config.Auth.AccessToken {
			sess.Delete("auth_token")
			sess.Delete("authenticated")
			if err := sess.Save(); err != nil {
				m.logger.WithError(err).Warn("Gagal menghapus sesi invalid")
			}
			return c.Redirect("/login?error=invalid_session&redirect=" + c.Path())
		}

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
