package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

var (
	ErrInvalidToken     = errors.New("token tidak valid")
	ErrExpiredToken     = errors.New("token telah kedaluwarsa")
	ErrTokenNotProvided = errors.New("token tidak disediakan")
)

// TokenManager bertanggung jawab untuk mengelola token akses
type TokenManager struct {
	config *config.AuthConfig
	logger utils.LogrusEntry
}

// NewTokenManager membuat instance baru TokenManager
func NewTokenManager(cfg *config.AuthConfig, logger utils.LogrusEntry) *TokenManager {
	return &TokenManager{
		config: cfg,
		logger: logger.WithField("component", "token-manager"),
	}
}

// TokenClaims mendefinisikan struktur claim JWT
type TokenClaims struct {
	jwt.RegisteredClaims
	TokenID   string `json:"token_id"`
	TokenType string `json:"token_type"`
}

// GenerateAccessToken menghasilkan token JWT baru dengan masa aktif yang ditentukan
func (tm *TokenManager) GenerateAccessToken(identifier string, duration time.Duration) (string, error) {
	// Buat token ID unik
	tokenID, err := generateRandomString(24)
	if err != nil {
		return "", fmt.Errorf("gagal membuat token ID: %w", err)
	}

	now := time.Now()
	expiryTime := now.Add(duration)

	// Buat klaim token
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identifier,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			ID:        tokenID,
		},
		TokenID:   tokenID,
		TokenType: "access",
	}

	// Buat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tangani token dengan secret
	if tm.config.TokenSecret == "" {
		return "", errors.New("token secret tidak boleh kosong")
	}

	signedToken, err := token.SignedString([]byte(tm.config.TokenSecret))
	if err != nil {
		return "", fmt.Errorf("gagal menandatangani token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken memvalidasi token akses
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	// Parse token dengan metode validasi
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validasi algoritma tanda tangan
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tidak valid: %v", token.Header["alg"])
		}
		return []byte(tm.config.TokenSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Validasi klaim token
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validasi tambahan
	if claims.TokenType != "access" {
		return nil, errors.New("tipe token tidak valid")
	}

	return claims, nil
}

// ValidateAPIToken memvalidasi token API statis
func (tm *TokenManager) ValidateAPIToken(token string) bool {
	if token == "" {
		return false
	}

	// Direct comparison - hanya untuk development
	if tm.config.AccessToken != "" && token == tm.config.AccessToken {
		return true
	}

	// Hash token yang diberikan dan bandingkan dengan yang tersimpan
	hashedToken := hashToken(token)

	// Jika hashed token tersedia, bandingkan
	if tm.config.HashedAccessToken != "" {
		return tm.config.HashedAccessToken == hashedToken
	}

	return false
}

// VerifyTokenHash memverifikasi apakah token cocok dengan hash yang disimpan
func (tm *TokenManager) VerifyTokenHash(token, hashedToken string) bool {
	hashed := hashToken(token)
	return hashed == hashedToken
}

// SetAPIToken menyimpan token API baru dengan hash
func (tm *TokenManager) SetAPIToken(token string) (string, error) {
	if len(token) < 24 {
		return "", errors.New("token terlalu pendek, minimal 24 karakter")
	}

	// Hash token
	hashedToken := hashToken(token)

	// Return hash untuk disimpan di konfigurasi
	return hashedToken, nil
}

// GenerateRandomToken menghasilkan token acak untuk API key
func (tm *TokenManager) GenerateRandomToken(length int) (string, error) {
	if length < 24 {
		length = 24 // Minimum length
	}

	token, err := generateRandomString(length)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateRandomString menghasilkan string acak dengan panjang tertentu
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// hashToken menghash token menggunakan SHA-256
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
