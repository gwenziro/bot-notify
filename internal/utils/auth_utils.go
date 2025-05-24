package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GenerateRandomToken menghasilkan token acak dengan panjang tertentu
func GenerateRandomToken(length int) (string, error) {
	if length < 24 {
		length = 24 // Minimum length
	}

	// Hitung berapa byte yang diperlukan
	// Base64: 4 karakter = 3 byte
	requiredBytes := length * 3 / 4
	b := make([]byte, requiredBytes)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)

	// Potong token sesuai panjang yang diminta
	if len(token) > length {
		token = token[:length]
	}

	return token, nil
}

// GenerateSecurePassword menghasilkan password yang aman
func GenerateSecurePassword(length int) (string, error) {
	if length < 12 {
		length = 12 // Minimum length for security
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

// ExtractClientInfo mengambil informasi client dari request
func ExtractClientInfo(c *fiber.Ctx) map[string]string {
	return map[string]string{
		"ip":         c.IP(),
		"user_agent": c.Get("User-Agent"),
		"referer":    c.Get("Referer"),
		"host":       c.Hostname(),
		"protocol":   c.Protocol(),
		"method":     c.Method(),
		"path":       c.Path(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
}

// IsLocalIP memeriksa apakah IP lokal/private
func IsLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Periksa apakah IP private
	private := ip.IsPrivate()

	// Periksa apakah loopback
	private = private || ip.IsLoopback()

	return private
}

// MaskIP mengaburkan sebagian IP untuk logging
func MaskIP(ip string) string {
	if ip == "" {
		return ""
	}

	// Untuk IPv4
	if parts := strings.Split(ip, "."); len(parts) == 4 {
		return fmt.Sprintf("%s.%s.xxx.xxx", parts[0], parts[1])
	}

	// Untuk IPv6
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		return fmt.Sprintf("%s:%s:xxxx:xxxx:xxxx", parts[0], parts[1])
	}

	return ip
}

// GenerateCSRFToken menghasilkan token CSRF
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetSessionUserID mendapatkan ID user dari sesi
func GetSessionUserID(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		user, ok := c.Locals("user").(fiber.Map)
		if !ok {
			return "", false
		}

		id, ok := user["id"].(string)
		if !ok {
			return "", false
		}

		return id, true
	}

	id, ok := userID.(string)
	return id, ok
}

// IsAuthenticated memeriksa apakah request sudah terautentikasi
func IsAuthenticated(c *fiber.Ctx) bool {
	auth := c.Locals("authenticated")
	if auth == nil {
		return false
	}

	return auth.(bool)
}
