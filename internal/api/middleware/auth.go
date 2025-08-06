package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APIAuthMiddleware mengelola autentikasi untuk API dengan dukungan multi-user
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

// RequireAuth memvalidasi bahwa pengguna telah login untuk API dengan dukungan multi-user
func (m *APIAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Dapatkan token dari header API
		token := c.Get("X-Access-Token")

		// Jika tidak ada token di header, coba dari query parameter
		if token == "" {
			token = c.Query("access_token")
		}

		// Jika masih tidak ada token, coba dari session
		if token == "" {
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

		// Jika token masih kosong, kembalikan error 401
		if token == "" {
			m.logger.Warn("API token tidak ditemukan", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "API token tidak ditemukan",
				"code":    fiber.StatusUnauthorized,
			})
		}

		// Cek apakah ini adalah admin token
		configToken := m.config.Auth.AccessToken
		if strings.EqualFold(token, configToken) {
			// Ini adalah admin token
			c.Locals("userID", "admin")
			c.Locals("isAdmin", true)
			c.Locals("authenticated", true)
			c.Locals("auth_time", time.Now().Unix())
			
			m.logger.Debug("Admin token tervalidasi", utils.Fields{
				"path": c.Path(),
				"ip":   c.IP(),
			})
			
			return c.Next()
		}

		// Untuk sistem multi-user, token adalah userID
		// Validasi format userID (harus alphanumeric dan tidak kosong)
		userID := strings.TrimSpace(token)
		if !m.isValidUserID(userID) {
			m.logger.Warn("Format userID tidak valid", utils.Fields{
				"userID": userID,
				"path":   c.Path(),
				"ip":     c.IP(),
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Format user ID tidak valid",
				"code":    fiber.StatusUnauthorized,
			})
		}

		// Set userID di context untuk digunakan oleh handler
		c.Locals("userID", userID)
		c.Locals("isAdmin", false)
		c.Locals("authenticated", true)
		c.Locals("auth_time", time.Now().Unix())

		m.logger.Debug("User token tervalidasi", utils.Fields{
			"userID": userID,
			"path":   c.Path(),
			"ip":     c.IP(),
		})

		return c.Next()
	}
}

// isValidUserID memvalidasi format userID
func (m *APIAuthMiddleware) isValidUserID(userID string) bool {
	// UserID harus:
	// 1. Tidak kosong
	// 2. Panjang antara 3-50 karakter
	// 3. Hanya mengandung huruf, angka, underscore, dan dash
	if len(userID) < 3 || len(userID) > 50 {
		return false
	}

	for _, char := range userID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// GetUserIDFromContext mengambil userID dari fiber context
func GetUserIDFromContext(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("userID")
	if userID == nil {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// IsAdminFromContext memeriksa apakah user adalah admin
func IsAdminFromContext(c *fiber.Ctx) bool {
	isAdmin := c.Locals("isAdmin")
	if isAdmin == nil {
		return false
	}

	isAdminBool, ok := isAdmin.(bool)
	return ok && isAdminBool
}