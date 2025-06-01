package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APIAuthMiddleware mengelola autentikasi untuk API
type APIAuthMiddleware struct {
	config       *config.Config
	sessionStore *session.Store
	logger       utils.LogrusEntry
}

// NewAPIAuthMiddleware membuat instance baru APIAuthMiddleware
func NewAPIAuthMiddleware(cfg *config.Config, sessionStore *session.Store) *APIAuthMiddleware {
	return &APIAuthMiddleware{
		config:       cfg,
		sessionStore: sessionStore,
		logger:       utils.ForModule("api-middleware"),
	}
}

// RequireAuth memvalidasi bahwa pengguna telah login untuk API
func (m *APIAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Dapatkan token dari header API
		token := c.Get("X-Access-Token")

		// Jika tidak ada token di header, coba dari query parameter
		if token == "" {
			token = c.Query("access_token")
		}

		// Jika masih tidak ada token, coba dari session
		if token == "" && m.sessionStore != nil {
			sess, err := m.sessionStore.Get(c)
			if err == nil {
				authToken := sess.Get("auth_token")
				if authToken != nil {
					if authStr, ok := authToken.(string); ok {
						token = authStr
					}
				}
			}
		}

		// Tambahkan logging untuk debug
		m.logger.Debug("Validating API token", utils.Fields{
			"token_length": len(token),
			"has_token":    token != "",
			"path":         c.Path(),
			"headers":      c.GetReqHeaders(),
		})

		// Jika token masih kosong, kembalikan error 401
		if token == "" {
			// Cek jika request adalah AJAX atau API
			m.logger.Warn("API token tidak ditemukan", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})

			// PENTING: SELALU return JSON untuk endpoint API, tidak boleh redirect atau HTML
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "API token tidak ditemukan",
				"code":    fiber.StatusUnauthorized,
			})
		}

		// Validasi token
		configToken := m.config.Auth.AccessToken
		if !strings.EqualFold(token, configToken) {
			m.logger.Warn("API token tidak valid", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})

			// PENTING: SELALU return JSON untuk endpoint API, tidak boleh redirect atau HTML
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "API token tidak valid",
				"code":    fiber.StatusUnauthorized,
			})
		}

		// Set token yang valid di locals
		c.Locals("token", token)
		c.Locals("authenticated", true)
		c.Locals("auth_time", time.Now().Unix())

		return c.Next()
	}
}
