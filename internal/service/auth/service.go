package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ErrorMessages memetakan kode error ke pesan yang user-friendly
var ErrorMessages = map[string]string{
	"invalid_credentials": "Token akses tidak valid. Silakan coba lagi.",
	"invalid_session":     "Sesi Anda telah kedaluwarsa. Silakan login kembali.",
	"too_many_attempts":   "Terlalu banyak percobaan login gagal. Silakan coba lagi nanti.",
	"session_timeout":     "Sesi Anda telah kedaluwarsa karena tidak aktif terlalu lama.",
	"security_concern":    "Terdapat masalah keamanan dengan sesi Anda. Silakan login kembali.",
	"invalid_request":     "Permintaan tidak valid. Silakan coba lagi.",
}

// Service menangani logika autentikasi
type Service struct {
	config       *config.Config
	logger       utils.LogrusEntry
	sessionStore *session.Store
	storage      storage.Storage
	maxAttempts  int
}

// NewService membuat instance baru auth service
func NewService(cfg *config.Config, sessionStore *session.Store, logger utils.LogrusEntry) *Service {
	return &Service{
		config:       cfg,
		logger:       logger.WithField("component", "auth-service"),
		sessionStore: sessionStore,
		storage:      storage.GetStorage(),
		maxAttempts:  5, // Default ke 5 percobaan
	}
}

// IsAuthenticated memeriksa apakah pengguna sudah login
func (s *Service) IsAuthenticated(ctx *fiber.Ctx) bool {
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return false
	}

	authToken := sess.Get("auth_token")
	return authToken != nil && authToken.(string) == s.config.Auth.AccessToken
}

// GetErrorMessage menerjemahkan kode error ke pesan user-friendly
func (s *Service) GetErrorMessage(errorCode string) string {
	if msg, exists := ErrorMessages[errorCode]; exists {
		return msg
	}
	return ""
}

// GenerateCSRFToken membuat dan menyimpan CSRF token baru
func (s *Service) GenerateCSRFToken(ctx *fiber.Ctx) (string, error) {
	token := GenerateRandomToken(32)

	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return "", err
	}

	sess.Set("csrf_token", token)
	if err := sess.Save(); err != nil {
		return "", err
	}

	return token, nil
}

// VerifyCSRFToken memvalidasi CSRF token
func (s *Service) VerifyCSRFToken(ctx *fiber.Ctx, token string) bool {
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return false
	}

	storedToken := sess.Get("csrf_token")
	if storedToken == nil || token == "" {
		return false
	}

	isValid := token == storedToken

	// Hapus token setelah diverifikasi untuk mencegah reuse
	if isValid {
		sess.Delete("csrf_token")
		sess.Save()
	}

	return isValid
}

// CheckRateLimiting memeriksa rate limiting login berdasarkan IP
func (s *Service) CheckRateLimiting(ctx *fiber.Ctx) (bool, int, error) {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()

	attemptData, err := s.storage.Get(bgCtx, ipKey)

	attempts := 0
	if err == nil && len(attemptData) > 0 {
		attempts, _ = strconv.Atoi(string(attemptData))

		// Jika melebihi batas percobaan, tolak
		if attempts >= s.maxAttempts {
			return false, attempts, nil
		}
	}

	return true, attempts, nil
}

// IncrementLoginAttempt meningkatkan counter percobaan login
func (s *Service) IncrementLoginAttempt(ctx *fiber.Ctx) error {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()

	// Dapatkan nilai saat ini
	attemptData, err := s.storage.Get(bgCtx, ipKey)

	attempts := 1
	if err == nil && len(attemptData) > 0 {
		currentAttempts, _ := strconv.Atoi(string(attemptData))
		attempts = currentAttempts + 1
	}

	// Simpan nilai baru dengan TTL 15 menit
	return s.storage.SetWithTTL(bgCtx, ipKey, []byte(strconv.Itoa(attempts)), 15*time.Minute)
}

// ResetLoginAttempt mengatur ulang counter percobaan login
func (s *Service) ResetLoginAttempt(ctx *fiber.Ctx) error {
	ipKey := fmt.Sprintf("login_attempt:%s", ctx.IP())
	bgCtx := context.Background()

	return s.storage.Delete(bgCtx, ipKey)
}

// CreateSession membuat sesi login baru
func (s *Service) CreateSession(ctx *fiber.Ctx, token string, rememberMe bool) error {
	// Buat sesi
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return err
	}

	// Set informasi sesi
	sess.Set("auth_token", token)
	sess.Set("authenticated", true)
	sess.Set("last_activity_time", time.Now().Unix())

	// Simpan fingerprint perangkat untuk security
	deviceFingerprint := fmt.Sprintf("%s|%s", ctx.IP(), ctx.Get("User-Agent"))
	sess.Set("device_fingerprint", deviceFingerprint)

	if err := sess.Save(); err != nil {
		return err
	}

	// Set cookie untuk "remember me"
	if rememberMe {
		ctx.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    token,
			Path:     "/",
			MaxAge:   s.config.Auth.CookieMaxAge,
			Secure:   s.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	}

	return nil
}

// ClearSession menghapus sesi login
func (s *Service) ClearSession(ctx *fiber.Ctx) error {
	// Dapatkan dan hapus sesi
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return err
	}

	sess.Delete("auth_token")
	sess.Delete("authenticated")
	sess.Delete("last_activity_time")
	sess.Delete("device_fingerprint")

	if err := sess.Save(); err != nil {
		return err
	}

	// Hapus cookie auto-login
	ctx.Cookie(&fiber.Cookie{
		Name:     "auto_login",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   s.config.Server.BaseURL != "http://localhost:8080",
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return nil
}

// GetRedirectHTML mengembalikan HTML untuk redirect dengan localStorage
func (s *Service) GetRedirectHTML(token, redirectURL string) string {
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>Redirecting...</title></head>
	<body>
		<script>
			// Set token for API calls
			localStorage.setItem('access_token', '%s');
			sessionStorage.setItem('authenticated', 'true');
			// Redirect to requested page
			window.location.href = '%s';
		</script>
	</body>
	</html>
	`, token, redirectURL)
}

// GenerateRandomToken menghasilkan token acak dengan panjang tertentu
func GenerateRandomToken(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
