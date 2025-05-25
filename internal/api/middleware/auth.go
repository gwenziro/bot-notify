package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APIAuthMiddleware mengelola autentikasi untuk API endpoints
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
		logger:       utils.ForModule("api-auth-middleware"),
	}
}

// RequireAuth middleware untuk API endpoints yang mendukung multiple authentication methods
func (m *APIAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		m.logger.Debug("Checking API authentication", utils.Fields{
			"path":   c.Path(),
			"method": c.Method(),
			"ip":     c.IP(),
		})

		// Method 1: Check Authorization header (Bearer token)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if m.validateToken(token) {
					m.logger.Debug("Authentication successful via Bearer token", utils.Fields{
						"path": c.Path(),
						"ip":   c.IP(),
					})
					return c.Next()
				}
			}
		}

		// Method 2: Check X-Access-Token header
		accessToken := c.Get("X-Access-Token")
		if accessToken != "" {
			if m.validateToken(accessToken) {
				m.logger.Debug("Authentication successful via X-Access-Token", utils.Fields{
					"path": c.Path(),
					"ip":   c.IP(),
				})
				return c.Next()
			}
		}

		// Method 3: Check session (for web-based API calls)
		sess, err := m.sessionStore.Get(c)
		if err == nil {
			authenticated := sess.Get("authenticated")
			authToken := sess.Get("auth_token")

			if authenticated == true && authToken != nil {
				token, ok := authToken.(string)
				if ok && m.validateToken(token) {
					m.logger.Debug("Authentication successful via session", utils.Fields{
						"path": c.Path(),
						"ip":   c.IP(),
					})
					return c.Next()
				}
			}
		}

		// Method 4: Check cookie auto_login (fallback)
		autoLoginToken := c.Cookies("auto_login")
		if autoLoginToken != "" {
			if m.validateToken(autoLoginToken) {
				m.logger.Debug("Authentication successful via auto_login cookie", utils.Fields{
					"path": c.Path(),
					"ip":   c.IP(),
				})
				return c.Next()
			}
		}

		// All authentication methods failed
		m.logger.Warn("API access denied: All authentication methods failed", utils.Fields{
			"ip":               c.IP(),
			"path":             c.Path(),
			"has_auth_header":  authHeader != "",
			"has_access_token": accessToken != "",
			"has_session":      err == nil,
			"has_cookie":       autoLoginToken != "",
		})

		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
			"code":    401,
		})
	}
}

// validateToken memvalidasi token dengan secure comparison
func (m *APIAuthMiddleware) validateToken(token string) bool {
	if token == "" || m.config.Auth.AccessToken == "" {
		return false
	}

	// Secure string comparison untuk mencegah timing attacks
	return m.secureCompare(token, m.config.Auth.AccessToken)
}

// secureCompare membandingkan dua string dengan waktu konstan
func (m *APIAuthMiddleware) secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
