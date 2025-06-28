package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// UserManager mengelola instance bot WhatsApp untuk setiap pengguna
type UserManager struct {
	// clients menyimpan mapping userID -> client instance
	clients map[string]*client.Client
	
	// mutex untuk memastikan akses concurrent yang aman
	mutex sync.RWMutex
	
	// globalConfig adalah konfigurasi global yang akan digunakan sebagai template
	globalConfig *config.Config
	
	// logger untuk logging aktivitas user manager
	logger utils.LogrusEntry
	
	// ctx dan cancel untuk graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// UserClientInfo berisi informasi tentang client pengguna
type UserClientInfo struct {
	UserID      string    `json:"userId"`
	IsConnected bool      `json:"isConnected"`
	CreatedAt   time.Time `json:"createdAt"`
	LastActive  time.Time `json:"lastActive"`
}

// NewUserManager membuat instance baru UserManager
func NewUserManager(globalCfg *config.Config) *UserManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	logger := utils.ForModule("user-manager")
	
	manager := &UserManager{
		clients:      make(map[string]*client.Client),
		mutex:        sync.RWMutex{},
		globalConfig: globalCfg,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}
	
	// Mulai goroutine untuk cleanup periodic
	go manager.startPeriodicCleanup()
	
	logger.Info("UserManager berhasil diinisialisasi")
	
	return manager
}

// NewUserClient membuat instance client baru untuk pengguna tertentu
func (um *UserManager) NewUserClient(userID string) (*client.Client, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()
	
	// Cek apakah client sudah ada
	if existingClient, exists := um.clients[userID]; exists {
		um.logger.Info("Client sudah ada untuk pengguna", utils.Fields{
			"userID": userID,
		})
		return existingClient, nil
	}
	
	// Buat konfigurasi khusus untuk pengguna ini
	userConfig := um.createUserSpecificConfig(userID)
	
	// Pastikan direktori pengguna ada
	if err := um.ensureUserDirectories(userID); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori pengguna: %w", err)
	}
	
	// Buat client baru
	newClient, err := client.NewClient(userID, userConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat client untuk pengguna %s: %w", userID, err)
	}
	
	// Simpan client ke map
	um.clients[userID] = newClient
	
	um.logger.Info("Client baru berhasil dibuat", utils.Fields{
		"userID":      userID,
		"totalUsers":  len(um.clients),
	})
	
	return newClient, nil
}

// GetUserClient mengambil instance client untuk pengguna tertentu
func (um *UserManager) GetUserClient(userID string) (*client.Client, bool) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	
	client, exists := um.clients[userID]
	if exists {
		// Update last active time
		go um.updateLastActive(userID)
	}
	
	return client, exists
}

// RemoveUserClient menghapus dan membersihkan client pengguna
func (um *UserManager) RemoveUserClient(userID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()
	
	client, exists := um.clients[userID]
	if !exists {
		return fmt.Errorf("client tidak ditemukan untuk pengguna: %s", userID)
	}
	
	// Tutup koneksi client
	client.Disconnect()
	client.Close()
	
	// Hapus dari map
	delete(um.clients, userID)
	
	// Bersihkan direktori pengguna (opsional, bisa dikomentari jika ingin menyimpan data)
	if err := um.cleanupUserDirectories(userID); err != nil {
		um.logger.Warn("Gagal membersihkan direktori pengguna", utils.Fields{
			"userID": userID,
			"error":  err.Error(),
		})
	}
	
	um.logger.Info("Client pengguna berhasil dihapus", utils.Fields{
		"userID":     userID,
		"totalUsers": len(um.clients),
	})
	
	return nil
}

// GetAllUsers mengembalikan informasi semua pengguna yang aktif
func (um *UserManager) GetAllUsers() []UserClientInfo {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	
	users := make([]UserClientInfo, 0, len(um.clients))
	
	for userID, client := range um.clients {
		state, _ := client.GetConnectionStateSafe()
		
		users = append(users, UserClientInfo{
			UserID:      userID,
			IsConnected: state.IsConnected,
			CreatedAt:   state.ConnectedSince,
			LastActive:  state.LastActivity,
		})
	}
	
	return users
}

// GetUserCount mengembalikan jumlah pengguna yang aktif
func (um *UserManager) GetUserCount() int {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	
	return len(um.clients)
}

// ShutdownAll menutup semua client dan membersihkan resources
func (um *UserManager) ShutdownAll() {
	um.mutex.Lock()
	defer um.mutex.Unlock()
	
	um.logger.Info("Memulai shutdown semua client pengguna", utils.Fields{
		"totalUsers": len(um.clients),
	})
	
	// Cancel context untuk menghentikan goroutine
	um.cancel()
	
	// Tutup semua client
	for userID, client := range um.clients {
		um.logger.Info("Menutup client pengguna", utils.Fields{"userID": userID})
		
		client.Disconnect()
		client.Close()
	}
	
	// Bersihkan map
	um.clients = make(map[string]*client.Client)
	
	um.logger.Info("Semua client pengguna berhasil ditutup")
}

// createUserSpecificConfig membuat konfigurasi khusus untuk pengguna
func (um *UserManager) createUserSpecificConfig(userID string) *config.Config {
	// Clone konfigurasi global
	userConfig := *um.globalConfig
	
	// Modifikasi path untuk isolasi pengguna
	baseDataDir := filepath.Join(utils.ProjectRoot, "data", "users", userID)
	
	userConfig.WhatsApp.StoreDir = filepath.Join(baseDataDir, "whatsapp")
	userConfig.WhatsApp.QrCodeDir = filepath.Join(baseDataDir, "qrcodes")
	userConfig.Auth.SessionDir = filepath.Join(baseDataDir, "sessions")
	
	// Path storage juga perlu diisolasi
	userConfig.Storage.Path = filepath.Join(baseDataDir, "storage")
	
	return &userConfig
}

// ensureUserDirectories memastikan semua direktori pengguna ada
func (um *UserManager) ensureUserDirectories(userID string) error {
	baseDataDir := filepath.Join(utils.ProjectRoot, "data", "users", userID)
	
	dirs := []string{
		filepath.Join(baseDataDir, "whatsapp"),
		filepath.Join(baseDataDir, "qrcodes"),
		filepath.Join(baseDataDir, "sessions"),
		filepath.Join(baseDataDir, "storage"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("gagal membuat direktori %s: %w", dir, err)
		}
	}
	
	um.logger.Debug("Direktori pengguna berhasil dibuat", utils.Fields{
		"userID":  userID,
		"baseDir": baseDataDir,
	})
	
	return nil
}

// cleanupUserDirectories membersihkan direktori pengguna
func (um *UserManager) cleanupUserDirectories(userID string) error {
	baseDataDir := filepath.Join(utils.ProjectRoot, "data", "users", userID)
	
	if err := os.RemoveAll(baseDataDir); err != nil {
		return fmt.Errorf("gagal menghapus direktori pengguna %s: %w", userID, err)
	}
	
	um.logger.Debug("Direktori pengguna berhasil dibersihkan", utils.Fields{
		"userID":  userID,
		"baseDir": baseDataDir,
	})
	
	return nil
}

// updateLastActive memperbarui waktu aktivitas terakhir pengguna
func (um *UserManager) updateLastActive(userID string) {
	um.mutex.RLock()
	client, exists := um.clients[userID]
	um.mutex.RUnlock()
	
	if exists {
		client.UpdateLastActivity()
	}
}

// startPeriodicCleanup memulai goroutine untuk pembersihan periodik
func (um *UserManager) startPeriodicCleanup() {
	ticker := time.NewTicker(30 * time.Minute) // Cleanup setiap 30 menit
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			um.logger.Info("Menghentikan periodic cleanup")
			return
		case <-ticker.C:
			um.performCleanup()
		}
	}
}

// performCleanup melakukan pembersihan client yang tidak aktif
func (um *UserManager) performCleanup() {
	um.mutex.Lock()
	defer um.mutex.Unlock()
	
	inactiveThreshold := 2 * time.Hour // Client dianggap tidak aktif setelah 2 jam
	now := time.Now()
	
	var toRemove []string
	
	for userID, client := range um.clients {
		state, err := client.GetConnectionStateSafe()
		if err != nil {
			continue
		}
		
		// Cek apakah client tidak aktif dan tidak terhubung
		if !state.IsConnected && now.Sub(state.LastActivity) > inactiveThreshold {
			toRemove = append(toRemove, userID)
		}
	}
	
	// Hapus client yang tidak aktif
	for _, userID := range toRemove {
		um.logger.Info("Membersihkan client tidak aktif", utils.Fields{
			"userID": userID,
		})
		
		client := um.clients[userID]
		client.Close()
		delete(um.clients, userID)
		
		// Cleanup direktori (opsional)
		go um.cleanupUserDirectories(userID)
	}
	
	if len(toRemove) > 0 {
		um.logger.Info("Periodic cleanup selesai", utils.Fields{
			"removedUsers": len(toRemove),
			"activeUsers":  len(um.clients),
		})
	}
}

// IsUserActive memeriksa apakah pengguna memiliki client aktif
func (um *UserManager) IsUserActive(userID string) bool {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	
	client, exists := um.clients[userID]
	if !exists {
		return false
	}
	
	state, err := client.GetConnectionStateSafe()
	if err != nil {
		return false
	}
	
	return state.IsConnected
}

// GetUserConnectionState mendapatkan status koneksi pengguna
func (um *UserManager) GetUserConnectionState(userID string) (client.ConnectionState, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	
	userClient, exists := um.clients[userID]
	if !exists {
		return client.ConnectionState{}, fmt.Errorf("client tidak ditemukan untuk pengguna: %s", userID)
	}
	
	return userClient.GetConnectionStateSafe()
}