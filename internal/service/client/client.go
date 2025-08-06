package client

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/session"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"

	_ "modernc.org/sqlite"
)

// ClientStatus menunjukkan status koneksi WhatsApp
type ClientStatus string

const (
	StatusDisconnected ClientStatus = "disconnected"
	StatusConnecting   ClientStatus = "connecting"
	StatusConnected    ClientStatus = "connected"
	StatusLoggedOut    ClientStatus = "logged_out"
)

// ConnectionState menyimpan informasi status koneksi
type ConnectionState struct {
	Status         ClientStatus `json:"status"`
	IsConnected    bool         `json:"is_connected"`
	LastActivity   time.Time    `json:"last_activity"`
	Timestamp      time.Time    `json:"timestamp"`
	ConnectedSince time.Time    `json:"connected_since"`
}

// EventHandlerFunc adalah tipe fungsi untuk menangani event WhatsApp
type EventHandlerFunc func(interface{})

// Client adalah wrapper untuk klien WhatsApp dengan dukungan multi-user
type Client struct {
	// UserID adalah identifier unik untuk pengguna ini
	UserID string
	
	waClient        *whatsmeow.Client
	eventHandler    uint32
	deviceStore     *sqlstore.Container
	connectionState ConnectionState
	config          *config.WhatsAppConfig
	logger          utils.LogrusEntry
	qrChan          chan string
	SessionManager  *session.Manager

	// Tambahkan counter pesan dan mutex untuk mengamankan akses
	messagesSent  int64
	messagesMutex sync.RWMutex

	callbackHandlers map[string]func(interface{})
	ctx              context.Context
	cancel           context.CancelFunc
}

// GetConnectionState mengembalikan state koneksi saat ini
func (c *Client) GetConnectionState() ConnectionState {
	return c.connectionState
}

// SetConnectionState mengatur status koneksi saat ini
func (c *Client) SetConnectionState(status ClientStatus, isConnected bool) {
	previousStatus := c.connectionState.Status
	previousConnected := c.connectionState.IsConnected

	c.connectionState.Status = status
	c.connectionState.IsConnected = isConnected
	c.connectionState.Timestamp = time.Now()

	// Set ConnectedSince hanya jika berubah dari tidak terhubung menjadi terhubung
	if status == StatusConnected && isConnected && !previousConnected {
		c.connectionState.ConnectedSince = time.Now()
		c.logger.Info("Status berubah menjadi terhubung, setting ConnectedSince", utils.Fields{
			"userID":          c.UserID,
			"from_status":     previousStatus,
			"to_status":       status,
			"connected_since": c.connectionState.ConnectedSince.Format(time.RFC3339),
		})
	}

	// Reset ConnectedSince saat terputus
	if !isConnected && previousConnected {
		c.connectionState.ConnectedSince = time.Time{} // Set ke zero time
		c.logger.Info("Status berubah menjadi terputus, resetting ConnectedSince", utils.Fields{
			"userID": c.UserID,
		})
	}
}

// GetWhatsmeowClient mengembalikan referensi ke client whatsmeow
func (c *Client) GetWhatsmeowClient() *whatsmeow.Client {
	return c.waClient
}

// GetCallbackHandlers mengembalikan map callback handler yang terdaftar
func (c *Client) GetCallbackHandlers() map[string]func(interface{}) {
	return c.callbackHandlers
}

// RegisterCallback mendaftarkan callback untuk event tertentu
func (c *Client) RegisterCallback(eventName string, callback func(interface{})) {
	c.callbackHandlers[eventName] = callback
}

// UpdateLastActivity memperbarui waktu aktivitas terakhir
func (c *Client) UpdateLastActivity() {
	// Selalu perbarui LastActivity
	c.connectionState.LastActivity = time.Now()

	// Hanya jika ConnectedSince adalah zero time dan client terhubung,
	// Gunakan waktu sekarang sebagai ConnectedSince
	if c.connectionState.IsConnected && c.connectionState.ConnectedSince.IsZero() {
		c.connectionState.ConnectedSince = c.connectionState.LastActivity
		c.logger.Info("ConnectedSince diinisialisasi (zero time sebelumnya)", utils.Fields{
			"userID":          c.UserID,
			"connected_since": c.connectionState.ConnectedSince.Format(time.RFC3339),
			"last_activity":   c.connectionState.LastActivity.Format(time.RFC3339),
		})
	}
}

// GetSelfID mengembalikan JID dari perangkat sendiri
func (c *Client) GetSelfID() *types.JID {
	if c.waClient == nil || !c.waClient.IsLoggedIn() {
		return nil
	}

	return c.waClient.Store.ID
}

// IncrementMessageCount menambah counter pesan terkirim
func (c *Client) IncrementMessageCount() {
	c.messagesMutex.Lock()
	defer c.messagesMutex.Unlock()
	c.messagesSent++

	// Log increment untuk debug
	c.logger.Debug("Incremented message count", utils.Fields{
		"userID":        c.UserID,
		"current_count": c.messagesSent,
	})
}

// GetMessagesSent mengembalikan jumlah pesan yang berhasil dikirim
func (c *Client) GetMessagesSent() int64 {
	c.messagesMutex.RLock()
	defer c.messagesMutex.RUnlock()
	return c.messagesSent
}

// ResetMessageCount mengatur ulang counter pesan terkirim
func (c *Client) ResetMessageCount() {
	c.messagesMutex.Lock()
	defer c.messagesMutex.Unlock()

	// Log sebelum reset untuk debug
	c.logger.Debug("Resetting message count", utils.Fields{
		"userID":         c.UserID,
		"previous_count": c.messagesSent,
	})

	c.messagesSent = 0
}

// NewClient membuat instance baru dari klien WhatsApp untuk pengguna tertentu
func NewClient(userID string, cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Gunakan nama modul yang jelas untuk WhatsApp client dengan userID
	logger := utils.ForModule(fmt.Sprintf("client-%s", userID))

	// Make sure to call cancel() if we encounter an error
	defer func() {
		if r := recover(); r != nil {
			cancel() // Make sure to call cancel if we panic
			logger.Error("Panic saat membuat client", utils.Fields{
				"userID": userID,
				"error":  r,
			})
		}
	}()

	// Pastikan direktori penyimpanan ada (menggunakan path yang sudah diisolasi per user)
	if err := os.MkdirAll(cfg.WhatsApp.StoreDir, 0755); err != nil {
		cancel() // Call cancel to prevent context leak
		return nil, fmt.Errorf("gagal membuat direktori penyimpanan untuk user %s: %w", userID, err)
	}

	// Buat klien WhatsApp dengan database path yang spesifik untuk user
	dbPath := fmt.Sprintf("%s/store.db", cfg.WhatsApp.StoreDir)

	waLogger := NewWhatsmeowLogger(logger)

	deviceStore, err := sqlstore.New(ctx, "sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath), waLogger)
	if err != nil {
		cancel() // Call cancel to prevent context leak
		return nil, fmt.Errorf("gagal membuat device store untuk user %s: %w", userID, err)
	}

	// Create the client with dependencies injected
	client := &Client{
		UserID:           userID, // Set UserID untuk identifikasi
		deviceStore:      deviceStore,
		config:           &cfg.WhatsApp,
		logger:           logger,
		qrChan:           make(chan string, 1),
		callbackHandlers: make(map[string]func(interface{})),
		ctx:              ctx,
		cancel:           cancel,
		connectionState: ConnectionState{
			Status:       StatusDisconnected,
			IsConnected:  false,
			Timestamp:    time.Now(),
			LastActivity: time.Now(),
		},
	}

	// Create session manager dengan path yang sudah diisolasi per user
	client.SessionManager = session.NewManager(cfg, logger, deviceStore)

	logger.Info("Client WhatsApp berhasil dibuat untuk pengguna", utils.Fields{
		"userID":    userID,
		"storeDir":  cfg.WhatsApp.StoreDir,
		"qrCodeDir": cfg.WhatsApp.QrCodeDir,
	})

	return client, nil
}

// StartConnectionHealthCheck memulai goroutine untuk memeriksa koneksi secara berkala
func (c *Client) StartConnectionHealthCheck(checkInterval time.Duration) {
	if checkInterval == 0 {
		checkInterval = 5 * time.Minute // Default check setiap 5 menit
	}

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				// Context dibatalkan, hentikan health check
				c.logger.Info("Menghentikan health check koneksi WhatsApp", utils.Fields{
					"userID": c.UserID,
				})
				return
			case <-ticker.C:
				// Waktu untuk memeriksa koneksi
				c.logger.Debug("Melakukan health check koneksi WhatsApp", utils.Fields{
					"userID": c.UserID,
				})
				isConnected := c.PingConnection()

				if !isConnected && c.connectionState.IsConnected {
					c.logger.Warn("Koneksi terputus terdeteksi selama health check, mencoba koneksi ulang", utils.Fields{
						"userID": c.UserID,
					})

					// Set status ke disconnected
					c.SetConnectionState(StatusDisconnected, false)

					// Coba koneksi ulang
					err := c.Connect()
					if err != nil {
						c.logger.WithError(err).Error("Gagal menghubungkan ulang WhatsApp selama health check", utils.Fields{
							"userID": c.UserID,
						})
					} else {
						c.logger.Info("Berhasil menghubungkan ulang WhatsApp selama health check", utils.Fields{
							"userID": c.UserID,
						})
					}
				}
			}
		}
	}()

	c.logger.Info("Health check koneksi WhatsApp dimulai", utils.Fields{
		"userID":         c.UserID,
		"check_interval": checkInterval.String(),
	})
}