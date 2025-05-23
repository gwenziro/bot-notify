package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// Logger struct untuk logging API
type Logger struct {
	logger utils.LogrusEntry
}

// NewLogger membuat instance baru Logger middleware
func NewLogger(logger utils.LogrusEntry) *Logger {
	return &Logger{
		logger: logger.WithField("component", "logger-middleware"),
	}
}

// RequestLogger mencatat semua request API
func (m *Logger) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Catat request masuk
		m.logger.WithFields(utils.Fields{
			"method": c.Method(),
			"path":   c.Path(),
			"ip":     c.IP(),
		}).Info("API request diterima")

		// Lanjutkan ke handler
		err := c.Next()

		// Catat status response
		m.logger.WithFields(utils.Fields{
			"method": c.Method(),
			"path":   c.Path(),
			"status": c.Response().StatusCode(),
		}).Info("API request selesai")

		return err
	}
}
