package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// AuthMiddleware menangani autentikasi API
type AuthMiddleware struct {
	accessToken string
	logger      utils.LogrusEntry
}

// NewAuthMiddleware membuat instance baru AuthMiddleware
func NewAuthMiddleware(accessToken string) *AuthMiddleware {
	return &AuthMiddleware{
		accessToken: accessToken,
		logger:      utils.ForModule("auth-middleware"),
	}
}

// Validate memvalidasi token akses API
func (m *AuthMiddleware) Validate(c *fiber.Ctx) error {
	// Periksa token akses
	token := c.Get("X-Access-Token")
	if token == "" {
		token = c.Query("token")
	}

	if token == "" || (token != m.accessToken) {
		m.logger.Warn("Akses ditolak: Token tidak valid", utils.Fields{
			"path": c.Path(),
			"ip":   c.IP(),
		})

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"sukses": false,
			"pesan":  "Unauthorized: Token akses tidak valid atau tidak ada",
		})
	}

	return c.Next()
}
