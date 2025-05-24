package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// SecurityMiddleware menangani aspek keamanan aplikasi
type SecurityMiddleware struct {
	config *config.Config
	logger utils.LogrusEntry
}

// NewSecurityMiddleware membuat instance baru SecurityMiddleware
func NewSecurityMiddleware(cfg *config.Config) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: cfg,
		logger: utils.ForModule("security-middleware"),
	}
}

// SecureHeaders menambahkan header keamanan ke respons
func (m *SecurityMiddleware) SecureHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		// Content-Security-Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' cdnjs.cloudflare.com; " +
			"style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com fonts.googleapis.com; " +
			"font-src 'self' fonts.gstatic.com cdnjs.cloudflare.com; " +
			"img-src 'self' data: https://img.icons8.com; " +
			"connect-src 'self'; " +
			"frame-src 'none'; " +
			"object-src 'none';"
		c.Set("Content-Security-Policy", csp)

		// Strict-Transport-Security for HTTPS
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Add server timing for debugging (can be removed in production)
		// Menggunakan property Debug yang berada di Server Config
		if m.config.Server.Debug {
			c.Set("Server-Timing", fmt.Sprintf("app;dur=%d", time.Now().UnixNano()))
		}

		// Lanjutkan ke handler berikutnya
		return c.Next()
	}
}

// RateLimiter membatasi jumlah request dalam periode tertentu
func (m *SecurityMiddleware) RateLimiter() fiber.Handler {
	// SimpleRateLimiter untuk demonstrasi
	// Untuk produksi, gunakan library seperti github.com/gofiber/fiber/v2/middleware/limiter
	// dengan storage persisten

	type rateLimitEntry struct {
		Count     int       `json:"count"`
		StartTime time.Time `json:"start_time"`
	}

	// In-memory map untuk contoh
	// Untuk produksi gunakan Redis atau storage persisten
	limitMap := make(map[string]*rateLimitEntry)

	return func(c *fiber.Ctx) error {
		// Gunakan IP sebagai kunci
		ip := c.IP()

		// Kurangi logging dengan hanya mencatat ratelimit ketika terlewati
		defer func() {
			// Jika handler diakhiri, bersihkan map untuk mencegah memory leak
			// Untuk produksi, gunakan mekanisme expiry/TTL
			if len(limitMap) > 1000 {
				// Bersihkan entri lama
				now := time.Now()
				for k, v := range limitMap {
					if now.Sub(v.StartTime) > 10*time.Minute {
						delete(limitMap, k)
					}
				}
			}
		}()

		// Batas maksimum untuk API dan web berbeda
		var maxRequests int
		var windowSize time.Duration

		// Path dimulai dengan /api dikenakan ratelimit yang berbeda
		if c.Path()[:4] == "/api" {
			maxRequests = 60 // 60 request per menit untuk API
			windowSize = 1 * time.Minute
		} else {
			maxRequests = 120 // 120 request per menit untuk web
			windowSize = 1 * time.Minute
		}

		// Periksa dan perbarui rate limit
		entry, exists := limitMap[ip]
		now := time.Now()

		if !exists || now.Sub(entry.StartTime) > windowSize {
			// Reset time window jika baru atau window sebelumnya sudah berlalu
			limitMap[ip] = &rateLimitEntry{
				Count:     1,
				StartTime: now,
			}
		} else {
			// Perbarui counter
			entry.Count++

			// Jika melebihi batas maksimum, tolak request
			if entry.Count > maxRequests {
				m.logger.WithFields(utils.Fields{
					"ip":     utils.MaskIP(ip),
					"path":   c.Path(),
					"method": c.Method(),
					"count":  entry.Count,
					"limit":  maxRequests,
					"window": windowSize.String(),
				}).Warn("Rate limit exceeded")

				// Header rate limit yang informatif
				c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
				c.Set("X-RateLimit-Remaining", "0")
				c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", entry.StartTime.Add(windowSize).Unix()))

				// Response untuk client tentang rate limiting
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"success": false,
					"error":   "Terlalu banyak permintaan, coba lagi nanti",
					"code":    "RATE_LIMIT_EXCEEDED",
				})
			}
		}

		// Set header rate limit
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-limitMap[ip].Count))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", entry.StartTime.Add(windowSize).Unix()))

		return c.Next()
	}
}

// CSRF protection middleware
func (m *SecurityMiddleware) CSRF() fiber.Handler {
	// Implementasi sederhana untuk demonstrasi
	// Untuk produksi, gunakan library seperti github.com/gofiber/fiber/v2/middleware/csrf

	return func(c *fiber.Ctx) error {
		// Skip untuk GET, HEAD, OPTIONS
		if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
			// Generate CSRF token untuk form
			token, err := utils.GenerateRandomToken(32)
			if err == nil {
				c.Locals("csrf_token", token)
			}
			return c.Next()
		}

		// Untuk POST dan lainnya, validasi CSRF token
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "DELETE" || c.Method() == "PATCH" {
			// Ambil token dari header atau form
			token := c.Get("X-CSRF-Token")
			if token == "" {
				token = c.FormValue("csrf_token")
			}

			expectedToken := c.Locals("csrf_token")

			// Jika token tidak ada, tolak
			if token == "" || expectedToken == nil || token != expectedToken.(string) {
				m.logger.WithFields(utils.Fields{
					"ip":     utils.MaskIP(c.IP()),
					"path":   c.Path(),
					"method": c.Method(),
				}).Warn("CSRF token tidak valid")

				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"error":   "CSRF token tidak valid",
					"code":    "INVALID_CSRF_TOKEN",
				})
			}
		}

		return c.Next()
	}
}
