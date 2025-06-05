package website

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// AuthService menyediakan fungsionalitas untuk autentikasi
type AuthService struct {
	config       *config.Config
	sessionStore *session.Store
	logger       utils.LogrusEntry
	store        storage.Storage
}

// NewAuthService membuat instance service autentikasi baru
func NewAuthService(cfg *config.Config, sessionStore *session.Store, logger utils.LogrusEntry) *AuthService {
	return &AuthService{
		config:       cfg,
		sessionStore: sessionStore,
		logger:       logger.WithField("component", "auth-service"),
		store:        storage.GetStorage(),
	}
}

// ValidateCredentials memeriksa apakah token yang diberikan valid
func (s *AuthService) ValidateCredentials(token string) bool {
	return token == s.config.Auth.AccessToken
}

// ValidateRateLimit memeriksa apakah IP melakukan terlalu banyak percobaan login
func (s *AuthService) ValidateRateLimit(ip string) bool {
	ipKey := fmt.Sprintf("login_attempt:%s", ip)
	bgCtx := context.Background()

	attemptData, err := s.store.Get(bgCtx, ipKey)
	if err != nil {
		return true // Asumsikan valid jika gagal mendapatkan data
	}

	if len(attemptData) > 0 {
		attempts, _ := strconv.Atoi(string(attemptData))
		if attempts >= 5 {
			s.logger.Warn("Login rate limited", utils.Fields{
				"ip":       ip,
				"attempts": attempts,
			})
			return false
		}
	}

	return true
}

// IncrementLoginAttempt menambahkan hitungan percobaan login gagal
func (s *AuthService) IncrementLoginAttempt(ip string) {
	ipKey := fmt.Sprintf("login_attempt:%s", ip)
	bgCtx := context.Background()

	attemptData, err := s.store.Get(bgCtx, ipKey)
	attempts := 1
	if err == nil && len(attemptData) > 0 {
		attempts, _ = strconv.Atoi(string(attemptData))
		attempts++
	}

	err = s.store.Set(bgCtx, ipKey, []byte(strconv.Itoa(attempts)))
	if err != nil {
		s.logger.WithError(err).Warn("Gagal menyimpan percobaan login")
	}
}

// ResetLoginAttempt menghapus hitungan percobaan login
func (s *AuthService) ResetLoginAttempt(ip string) {
	ipKey := fmt.Sprintf("login_attempt:%s", ip)
	bgCtx := context.Background()
	s.store.Delete(bgCtx, ipKey)
}

// CreateSession membuat sesi login baru
func (s *AuthService) CreateSession(c *fiber.Ctx, token string) error {
	sess, err := s.sessionStore.Get(c)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	// Set token dan status autentikasi di sesi
	sess.Set("auth_token", token)
	sess.Set("authenticated", true)
	sess.Set("last_activity_time", time.Now().Unix())

	// Simpan fingerprint perangkat
	deviceFingerprint := fmt.Sprintf("%s|%s", c.IP(), c.Get("User-Agent"))
	sess.Set("device_fingerprint", deviceFingerprint)

	return sess.Save()
}

// SetAutoLoginCookie menetapkan cookie untuk auto-login
func (s *AuthService) SetAutoLoginCookie(c *fiber.Ctx, enabled bool) {
	if enabled {
		c.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    s.config.Auth.AccessToken,
			Path:     "/",
			MaxAge:   s.config.Auth.CookieMaxAge,
			Secure:   s.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	} else {
		c.Cookie(&fiber.Cookie{
			Name:     "auto_login",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   s.config.Server.BaseURL != "http://localhost:8080",
			HTTPOnly: true,
			SameSite: "Strict",
		})
	}
}

// ClearSession membersihkan sesi saat logout
func (s *AuthService) ClearSession(c *fiber.Ctx) error {
	sess, err := s.sessionStore.Get(c)
	if err != nil {
		return err
	}

	sess.Delete("auth_token")
	sess.Delete("authenticated")
	sess.Delete("last_activity_time")
	sess.Delete("device_fingerprint")

	return sess.Save()
}

// VerifyCsrfToken memeriksa validitas token CSRF
func (s *AuthService) VerifyCsrfToken(c *fiber.Ctx, csrfToken string) bool {
	sess, err := s.sessionStore.Get(c)
	if err != nil {
		s.logger.WithError(err).Error("Gagal mendapatkan sesi saat validasi CSRF")
		return false
	}

	storedToken := sess.Get("csrf_token")
	if storedToken == nil || csrfToken == "" || csrfToken != storedToken {
		s.logger.Warn("CSRF token tidak valid", utils.Fields{
			"ip": c.IP(),
		})
		return false
	}

	// Hapus CSRF token setelah digunakan
	sess.Delete("csrf_token")
	sess.Save()

	return true
}

// GetMaskedToken mengembalikan token akses yang sudah disamarkan
func (s *AuthService) GetMaskedToken() string {
	return utils.MaskToken(s.config.Auth.AccessToken)
}

// GetErrorMessage menerjemahkan kode error menjadi pesan yang bisa dibaca
func (s *AuthService) GetErrorMessage(errorCode string) string {
	switch errorCode {
	case "invalid_credentials":
		return "Token akses tidak valid. Silakan coba lagi."
	case "invalid_session":
		return "Sesi Anda telah kedaluwarsa. Silakan login kembali."
	case "too_many_attempts":
		return "Terlalu banyak percobaan login gagal. Silakan coba lagi nanti."
	case "session_timeout":
		return "Sesi Anda telah kedaluwarsa karena tidak aktif terlalu lama."
	case "security_concern":
		return "Terdapat masalah keamanan dengan sesi Anda. Silakan login kembali."
	case "invalid_request":
		return "Permintaan tidak valid. Silakan coba lagi."
	default:
		return ""
	}
}

// IsLoggedIn memeriksa apakah pengguna sudah login
func (s *AuthService) IsLoggedIn(ctx *fiber.Ctx) (bool, error) {
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return false, err
	}

	authToken := sess.Get("auth_token")
	authenticated := sess.Get("authenticated")

	return authenticated != nil && authenticated.(bool) && authToken != nil && authToken.(string) == s.config.Auth.AccessToken, nil
}

// GenerateCSRFToken membuat dan menyimpan token CSRF baru
func (s *AuthService) GenerateCSRFToken(ctx *fiber.Ctx) (string, error) {
	csrfToken := utils.GenerateRandomToken(32)
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return "", err
	}

	sess.Set("csrf_token", csrfToken)
	if err := sess.Save(); err != nil {
		return "", err
	}

	return csrfToken, nil
}

// SaveCSRFToken menyimpan token CSRF ke sesi
func (s *AuthService) SaveCSRFToken(ctx *fiber.Ctx, csrfToken string) error {
	sess, err := s.sessionStore.Get(ctx)
	if err != nil {
		return err
	}

	sess.Set("csrf_token", csrfToken)
	return sess.Save()
}

// CreateLoginResponse membuat respons setelah login berhasil
func (s *AuthService) CreateLoginResponse(ctx *fiber.Ctx, token string, redirect string) error {
	redirectHTML := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Redirecting...</title>
		<script>
			// Fungsi untuk memastikan token disimpan dengan benar
			function ensureTokenSaved(token) {
				localStorage.setItem('access_token', token);
				
				// Verifikasi token tersimpan dengan benar
				const savedToken = localStorage.getItem('access_token');
				if (savedToken !== token) {
					console.error('Token tidak tersimpan dengan benar di localStorage');
					alert('Terjadi masalah saat menyimpan token autentikasi. Mohon reload halaman jika mengalami masalah.');
				} else {
					console.log('Token berhasil disimpan di localStorage');
				}
				
				sessionStorage.setItem('authenticated', 'true');
			}
			
			// Simpan token untuk API calls
			ensureTokenSaved('%s');
			
			// Redirect ke halaman yang diminta
			window.location.href = '%s';
		</script>
	</head>
	<body>
		<p>Redirecting to dashboard...</p>
	</body>
	</html>
	`, token, redirect)

	ctx.Set("Content-Type", "text/html")
	return ctx.SendString(redirectHTML)
}

// PrepareLoginPage menyiapkan semua data yang diperlukan untuk halaman login
func (s *AuthService) PrepareLoginPage(ctx *fiber.Ctx, redirect, errorCode string) (map[string]interface{}, error) {
	// Generate CSRF token
	csrfToken := utils.GenerateRandomToken(32)

	// Simpan CSRF token ke session
	if err := s.SaveCSRFToken(ctx, csrfToken); err != nil {
		return nil, err
	}

	// Siapkan data untuk template
	return fiber.Map{
		"Title":       "Login - WhatsApp Bot Notify",
		"RedirectTo":  redirect,
		"Error":       s.GetErrorMessage(errorCode),
		"CsrfToken":   csrfToken,
		"AccessToken": s.GetMaskedToken(),
	}, nil
}

// ProcessLogin memproses login dan menangani seluruh logika autentikasi
func (s *AuthService) ProcessLogin(ctx *fiber.Ctx, token string, rememberMe bool, redirect, csrfToken string) (bool, string, error) {
	// Buat entity request untuk pemrosesan internal
	req := entity.LoginRequest{
		Token:      token,
		RememberMe: rememberMe,
		CsrfToken:  csrfToken,
	}

	// Verifikasi CSRF token
	if !s.VerifyCsrfToken(ctx, req.CsrfToken) {
		s.logger.Warn("CSRF token tidak valid", utils.Fields{
			"ip": ctx.IP(),
		})
		return false, "invalid_request", nil
	}

	// Validasi rate limiting
	if !s.ValidateRateLimit(ctx.IP()) {
		s.logger.Warn("Rate limit tercapai", utils.Fields{
			"ip": ctx.IP(),
		})
		return false, "too_many_attempts", nil
	}

	// Validasi kredensial
	if !s.ValidateCredentials(req.Token) {
		// Tambah counter percobaan login gagal
		s.IncrementLoginAttempt(ctx.IP())
		s.logger.Warn("Percobaan login gagal", utils.Fields{
			"ip": ctx.IP(),
		})
		return false, "invalid_credentials", nil
	}

	// Reset counter percobaan login
	s.ResetLoginAttempt(ctx.IP())

	// Buat sesi
	if err := s.CreateSession(ctx, req.Token); err != nil {
		s.logger.WithError(err).Error("Gagal membuat sesi login")
		return false, "session_error", err
	}

	// Set cookie auto-login jika "remember me" dicentang
	if req.RememberMe {
		s.SetAutoLoginCookie(ctx, true)
	}

	s.logger.Info("Login berhasil", utils.Fields{
		"ip":          ctx.IP(),
		"remember_me": req.RememberMe,
	})

	return true, "", nil
}

// HandleLogout menangani seluruh proses logout
func (s *AuthService) HandleLogout(ctx *fiber.Ctx) error {
	// Hapus sesi
	if err := s.ClearSession(ctx); err != nil {
		s.logger.WithError(err).Warn("Gagal menghapus sesi saat logout")
	}

	// Hapus cookie auto-login
	s.SetAutoLoginCookie(ctx, false)

	s.logger.Info("User logged out", utils.Fields{
		"ip": ctx.IP(),
	})

	return nil
}
