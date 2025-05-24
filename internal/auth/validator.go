package auth

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("token akses tidak valid")
	ErrAuthRequired       = errors.New("autentikasi diperlukan")
)

// AuthValidator menangani validasi autentikasi
type AuthValidator struct {
	config         *config.AuthConfig
	tokenManager   *TokenManager
	sessionManager *SessionManager
	logger         utils.LogrusEntry
}

// NewAuthValidator membuat instance baru AuthValidator
func NewAuthValidator(config *config.AuthConfig, tokenManager *TokenManager, sessionManager *SessionManager, logger utils.LogrusEntry) *AuthValidator {
	return &AuthValidator{
		config:         config,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
		logger:         logger.WithField("component", "auth-validator"),
	}
}

// ValidateAPIAccess memvalidasi akses API menggunakan token di header
func (av *AuthValidator) ValidateAPIAccess(c *fiber.Ctx) error {
	// Dapatkan token dari header
	token := c.Get("X-Access-Token")
	if token == "" {
		// Coba dari header Authorization
		authHeader := c.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	// Jika masih tidak ada token, coba query param (kurang aman, tapi didukung untuk kompatibilitas)
	if token == "" {
		token = c.Query("access_token")
	}

	// Jika masih tidak ada token, return error
	if token == "" {
		return ErrTokenNotProvided
	}

	// Validasi token
	if av.tokenManager.ValidateAPIToken(token) {
		return nil
	}

	av.logger.WithField("ip", c.IP()).Warn("Percobaan akses API dengan token tidak valid")
	return ErrInvalidCredentials
}

// ValidateLogin memvalidasi token login
func (av *AuthValidator) ValidateLogin(inputToken string) error {
	if inputToken == "" {
		return ErrTokenNotProvided
	}

	// Bandingkan dengan token yang tersimpan (menggunakan constant-time comparison)
	configToken := av.config.AccessToken
	if subtle.ConstantTimeCompare([]byte(inputToken), []byte(configToken)) != 1 {
		return ErrInvalidCredentials
	}

	return nil
}

// ValidateWebAccess memvalidasi akses web menggunakan session
func (av *AuthValidator) ValidateWebAccess(c *fiber.Ctx) error {
	// Periksa apakah pengguna terautentikasi melalui sesi
	if !av.sessionManager.IsAuthenticated(c) {
		return ErrAuthRequired
	}

	// Perbarui timestamp aktivitas untuk sesi
	err := av.sessionManager.UpdateSessionActivity(c)
	if err != nil {
		av.logger.WithError(err).Error("Gagal memperbarui aktivitas sesi")
		// Tidak return error di sini, karena kita masih ingin melanjutkan jika pengguna terautentikasi
	}

	return nil
}

// CheckPasswordBruteForce memeriksa dan mencatat percobaan login yang gagal
func (av *AuthValidator) CheckPasswordBruteForce(c *fiber.Ctx) (bool, error) {
	// Implementasi anti-brute force bisa ditambahkan di sini
	// Misalnya: rate-limiting berdasarkan IP, cooldown setelah X percobaan gagal, dll.
	// Untuk sekarang kita hanya mencatat percobaan gagal

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	av.logger.WithFields(utils.Fields{
		"ip":         ip,
		"user_agent": userAgent,
		"timestamp":  time.Now().Format(time.RFC3339),
	}).Warn("Percobaan login gagal")

	// TODO: Tambahkan implementasi brute force protection
	// Contoh: jika percobaan gagal > X dalam Y menit, blokir untuk Z menit

	return false, nil
}

// GetAuthRedirectURL mengembalikan URL untuk redirect setelah auth
func (av *AuthValidator) GetAuthRedirectURL(c *fiber.Ctx, defaultPath string) string {
	// Dapatkan URL redirect dari query param atau default ke dashboard
	redirectTo := c.Query("redirect", defaultPath)

	// Pastikan URL adalah path relatif untuk mencegah open redirect
	if redirectTo[0] != '/' {
		redirectTo = "/" + redirectTo
	}

	return redirectTo
}

// IsTokenExpired memeriksa apakah token sudah kedaluwarsa
func (av *AuthValidator) IsTokenExpired(token string) bool {
	_, err := av.tokenManager.ValidateAccessToken(token)
	return errors.Is(err, ErrExpiredToken)
}

// GenerateSecureToken menghasilkan token aman untuk API key
func (av *AuthValidator) GenerateSecureToken() (string, error) {
	return av.tokenManager.GenerateRandomToken(36)
}
