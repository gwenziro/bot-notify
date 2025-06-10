package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// APILoggerMiddleware mencatat semua aktivitas API
func APILoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Catat waktu mulai
		start := time.Now()

		// Simpan request method dan endpoint
		method := c.Method()
		endpoint := c.Path()

		// Jalankan handler
		err := c.Next()

		// Hitung durasi
		duration := time.Since(start)

		// Tentukan status log berdasarkan status code
		status := "success"
		if c.Response().StatusCode() >= 400 {
			status = "error"
		} else if c.Response().StatusCode() >= 300 {
			status = "warning"
		}

		// Log aktivitas API
		message := "API Request"
		if err != nil {
			message = err.Error()
		}

		// Log dengan fields tambahan
		utils.LogAPI(
			method,
			endpoint,
			status,
			message,
			duration,
			utils.Fields{
				"status_code": c.Response().StatusCode(),
				"ip":          c.IP(),
				"user_agent":  c.Get("User-Agent"),
				"request_id":  c.GetRespHeader("X-Request-ID", "-"),
			},
		)

		return err
	}
}
