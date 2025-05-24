package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ErrSessionNotFound menunjukkan bahwa sesi tidak ditemukan
var ErrSessionNotFound = errors.New("sesi tidak ditemukan")

// SessionData menyimpan data sesi yang perlu persisten
type SessionData struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	LoginTime  time.Time `json:"login_time"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
}

// SessionManager mengelola sesi pengguna
type SessionManager struct {
	store      *session.Store
	storage    storage.Storage
	config     *config.AuthConfig
	logger     utils.LogrusEntry
	storageKey string
}

// NewSessionManager membuat instance baru SessionManager
func NewSessionManager(store *session.Store, storage storage.Storage, cfg *config.AuthConfig, logger utils.LogrusEntry) *SessionManager {
	return &SessionManager{
		store:      store,
		storage:    storage,
		config:     cfg,
		logger:     logger.WithField("component", "session-manager"),
		storageKey: "sessions",
	}
}

// CreateSession membuat sesi baru dan menyimpannya ke storage
func (sm *SessionManager) CreateSession(c *fiber.Ctx, userID string, rememberMe bool) (*session.Session, error) {
	// Dapatkan sesi baru dari store
	sess, err := sm.store.Get(c)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	// Regenerasi ID sesi untuk mencegah session fixation
	if err := sess.Regenerate(); err != nil {
		return nil, fmt.Errorf("gagal meregenerasi sesi: %w", err)
	}

	// Set expiry sesuai konfigurasi
	expiry := sm.config.TokenExpiry
	if rememberMe {
		// Jika remember me, gunakan waktu kedaluwarsa yang lebih panjang (30 hari)
		expiry = 30 * 24 * time.Hour
	}

	// Simpan data pengguna dalam sesi
	sess.Set("user_id", userID)
	sess.Set("authenticated", true)
	sess.Set("login_time", time.Now().Unix())
	sess.Set("expires_at", time.Now().Add(expiry).Unix())

	// Buat data sesi untuk disimpan di storage
	sessionData := SessionData{
		UserID:     userID,
		Username:   "admin", // Default username
		LoginTime:  time.Now(),
		LastActive: time.Now(),
		ExpiresAt:  time.Now().Add(expiry),
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	// Simpan sesi ke storage untuk pelacakan
	if err := sm.saveSessionData(sess.ID(), &sessionData); err != nil {
		sm.logger.WithError(err).Error("Gagal menyimpan data sesi ke storage")
	}

	// Atur cookie expiry untuk sesi
	sess.SetExpiry(expiry)

	// Simpan sesi
	if err := sess.Save(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan sesi: %w", err)
	}

	return sess, nil
}

// GetSession mengambil sesi yang ada
func (sm *SessionManager) GetSession(c *fiber.Ctx) (*session.Session, error) {
	sess, err := sm.store.Get(c)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}
	return sess, nil
}

// DestroySession menghapus sesi yang ada
func (sm *SessionManager) DestroySession(c *fiber.Ctx) error {
	sess, err := sm.store.Get(c)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	// Hapus data sesi dari storage
	if err := sm.deleteSessionData(sess.ID()); err != nil {
		sm.logger.WithError(err).Error("Gagal menghapus data sesi dari storage")
	}

	// Hapus sesi dari store
	if err := sess.Destroy(); err != nil {
		return fmt.Errorf("gagal menghapus sesi: %w", err)
	}

	return nil
}

// UpdateSessionActivity memperbarui timestamp aktivitas terakhir
func (sm *SessionManager) UpdateSessionActivity(c *fiber.Ctx) error {
	sess, err := sm.store.Get(c)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	// Dapatkan data sesi dari storage
	sessionData, err := sm.getSessionData(sess.ID())
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			// Sesi mungkin tidak disimpan di storage, coba buat baru
			userID := sess.Get("user_id")
			if userID == nil {
				return errors.New("user ID tidak ditemukan di sesi")
			}

			// Buat data sesi baru
			sessionData = &SessionData{
				UserID:     userID.(string),
				Username:   "admin",
				LoginTime:  time.Now(),
				LastActive: time.Now(),
				ExpiresAt:  time.Now().Add(sm.config.TokenExpiry),
				IPAddress:  c.IP(),
				UserAgent:  c.Get("User-Agent"),
			}
		} else {
			return fmt.Errorf("gagal mendapatkan data sesi: %w", err)
		}
	}

	// Perbarui aktivitas terakhir
	sessionData.LastActive = time.Now()

	// Simpan kembali ke storage
	if err := sm.saveSessionData(sess.ID(), sessionData); err != nil {
		return fmt.Errorf("gagal memperbarui data sesi: %w", err)
	}

	return nil
}

// GetActiveSessions mengembalikan semua sesi aktif
func (sm *SessionManager) GetActiveSessions() ([]SessionData, error) {
	ctx := context.Background()

	// Dapatkan semua data sesi dari storage
	data, err := sm.storage.GetWithPrefix(ctx, sm.storageKey+":")
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan data sesi: %w", err)
	}

	var sessions []SessionData
	now := time.Now()

	// Proses setiap entri sesi
	for _, sessionBytes := range data {
		var sessionData SessionData
		if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
			sm.logger.WithError(err).Error("Gagal unmarshal data sesi")
			continue
		}

		// Periksa apakah sesi masih aktif
		if sessionData.ExpiresAt.After(now) {
			sessions = append(sessions, sessionData)
		}
	}

	return sessions, nil
}

// IsAuthenticated memeriksa apakah sesi saat ini terautentikasi
func (sm *SessionManager) IsAuthenticated(c *fiber.Ctx) bool {
	sess, err := sm.store.Get(c)
	if err != nil {
		return false
	}

	// Periksa flag otentikasi dalam sesi
	authenticated := sess.Get("authenticated")
	if authenticated == nil {
		return false
	}

	// Periksa apakah sesi sudah kedaluwarsa
	expiresAt := sess.Get("expires_at")
	if expiresAt != nil {
		expiry, ok := expiresAt.(int64)
		if ok && time.Now().Unix() > expiry {
			// Sesi kedaluwarsa, hapus sesi
			_ = sm.DestroySession(c)
			return false
		}
	}

	return authenticated.(bool)
}

// GetUserID mengambil user ID dari sesi
func (sm *SessionManager) GetUserID(c *fiber.Ctx) (string, error) {
	sess, err := sm.store.Get(c)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return "", errors.New("user ID tidak ditemukan di sesi")
	}

	return userID.(string), nil
}

// saveSessionData menyimpan data sesi ke storage
func (sm *SessionManager) saveSessionData(sessionID string, data *SessionData) error {
	ctx := context.Background()

	// Konversi data ke JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("gagal marshal data sesi: %w", err)
	}

	// Simpan ke storage dengan TTL
	ttl := time.Until(data.ExpiresAt)
	if ttl <= 0 {
		ttl = sm.config.TokenExpiry
	}

	err = sm.storage.SetWithTTL(ctx, sm.storageKey+":"+sessionID, jsonData, ttl)
	if err != nil {
		return fmt.Errorf("gagal menyimpan data sesi: %w", err)
	}

	return nil
}

// getSessionData mengambil data sesi dari storage
func (sm *SessionManager) getSessionData(sessionID string) (*SessionData, error) {
	ctx := context.Background()

	// Ambil data dari storage
	data, err := sm.storage.Get(ctx, sm.storageKey+":"+sessionID)
	if err != nil {
		if storage.IsNotFound(err) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("gagal mengambil data sesi: %w", err)
	}

	// Unmarshal data
	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("gagal unmarshal data sesi: %w", err)
	}

	return &sessionData, nil
}

// deleteSessionData menghapus data sesi dari storage
func (sm *SessionManager) deleteSessionData(sessionID string) error {
	ctx := context.Background()

	// Hapus data dari storage
	err := sm.storage.Delete(ctx, sm.storageKey+":"+sessionID)
	if err != nil && !storage.IsNotFound(err) {
		return fmt.Errorf("gagal menghapus data sesi: %w", err)
	}

	return nil
}

// ClearExpiredSessions membersihkan sesi yang kedaluwarsa
func (sm *SessionManager) ClearExpiredSessions() (int, error) {
	ctx := context.Background()

	// Dapatkan semua data sesi dari storage
	data, err := sm.storage.GetWithPrefix(ctx, sm.storageKey+":")
	if err != nil {
		return 0, fmt.Errorf("gagal mendapatkan data sesi: %w", err)
	}

	var deletedCount int
	now := time.Now()

	// Proses setiap entri sesi
	for key, sessionBytes := range data {
		var sessionData SessionData
		if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
			sm.logger.WithError(err).Error("Gagal unmarshal data sesi")
			continue
		}

		// Hapus sesi kedaluwarsa
		if sessionData.ExpiresAt.Before(now) {
			if err := sm.storage.Delete(ctx, key); err != nil {
				sm.logger.WithError(err).Error("Gagal menghapus sesi kedaluwarsa")
				continue
			}
			deletedCount++
		}
	}

	return deletedCount, nil
}
