package client

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/session"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"

	_ "modernc.org/sqlite"
)

// ClientStatus menunjukkan status koneksi WhatsApp
type ClientStatus string

// Konstanta status koneksi
const (
	StatusDisconnected ClientStatus = "disconnected" // Tidak terhubung
	StatusConnecting   ClientStatus = "connecting"   // Sedang menghubungkan
	StatusConnected    ClientStatus = "connected"    // Terhubung
	StatusLoggedOut    ClientStatus = "logged_out"   // Telah logout
)

// ConnectionState menyimpan informasi status koneksi
type ConnectionState struct {
	Status            ClientStatus // Status koneksi saat ini
	IsConnected       bool         // Flag apakah terhubung
	ConnectionRetries int          // Jumlah percobaan koneksi
	LastActivity      time.Time    // Waktu aktivitas terakhir
	Timestamp         time.Time    // Timestamp status terakhir
	ConnectedSince    time.Time    // Waktu terhubung pertama kali
}

// Client adalah wrapper untuk klien WhatsApp
type Client struct {
	// Dependensi eksternal
	waClient       *whatsmeow.Client      // Klien whatsmeow
	deviceStore    *sqlstore.Container    // Container penyimpanan perangkat
	SessionManager *session.Manager       // Pengelola sesi WhatsApp
	config         *config.WhatsAppConfig // Konfigurasi WhatsApp
	logger         utils.LogrusEntry      // Logger untuk pencatatan aktivitas

	// State internal
	connectionState  ConnectionState              // Status koneksi saat ini
	callbackHandlers map[string]func(interface{}) // Callback untuk event
	qrChan           chan string                  // Channel untuk QR code

	// Kontrol konkurensi
	reconnectLock sync.Mutex         // Lock untuk operasi reconnect
	retryTimer    *time.Timer        // Timer untuk percobaan ulang
	ctx           context.Context    // Konteks utama
	cancel        context.CancelFunc // Fungsi untuk membatalkan konteks
}

// NewClient membuat instance baru dari klien WhatsApp
// Parameters:
// - cfg: konfigurasi aplikasi
// Returns:
// - klien WhatsApp
// - error jika gagal
func NewClient(cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Gunakan nama modul yang jelas untuk WhatsApp client
	logger := utils.ForModule("client")

	// Make sure to call cancel() if we encounter an error
	defer func() {
		if r := recover(); r != nil {
			cancel() // Make sure to call cancel if we panic
			logger.Error("Panic saat membuat client", utils.Fields{"error": r})
		}
	}()

	// Pastikan direktori penyimpanan ada
	if err := os.MkdirAll(cfg.WhatsApp.StoreDir, 0755); err != nil {
		cancel() // Call cancel to prevent context leak
		return nil, fmt.Errorf("gagal membuat direktori penyimpanan: %w", err)
	}

	// Buat klien WhatsApp
	dbPath := fmt.Sprintf("%s/store.db", cfg.WhatsApp.StoreDir)

	waLogger := NewWhatsmeowLogger(logger)

	deviceStore, err := sqlstore.New(ctx, "sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath), waLogger)
	if err != nil {
		cancel() // Call cancel to prevent context leak
		return nil, fmt.Errorf("gagal membuat device store: %w", err)
	}

	// Create the client with dependencies injected
	client := &Client{
		deviceStore:      deviceStore,
		config:           &cfg.WhatsApp,
		logger:           logger,
		qrChan:           make(chan string, 1),
		callbackHandlers: make(map[string]func(interface{})),
		ctx:              ctx,
		cancel:           cancel,
		connectionState: ConnectionState{
			Status:            StatusDisconnected,
			IsConnected:       false,
			ConnectionRetries: 0,
			Timestamp:         time.Now(),
			LastActivity:      time.Now(),
		},
		reconnectLock: sync.Mutex{},
	}

	// Create session manager with callback
	client.SessionManager = session.NewManager(cfg, logger, deviceStore)

	return client, nil
}

// GetConnectionState mengembalikan state koneksi saat ini
// Returns: salinan dari state koneksi
func (c *Client) GetConnectionState() ConnectionState {
	return c.connectionState
}

// SetConnectionState mengatur status koneksi saat ini
// Parameters:
// - status: status koneksi baru
// - isConnected: flag apakah terhubung
// - retries: jumlah percobaan koneksi
func (c *Client) SetConnectionState(status ClientStatus, isConnected bool, retries int) {
	previousStatus := c.connectionState.Status

	c.connectionState.Status = status
	c.connectionState.IsConnected = isConnected
	c.connectionState.ConnectionRetries = retries
	c.connectionState.Timestamp = time.Now()

	// Set ConnectedSince hanya jika status berubah dari tidak terhubung menjadi terhubung
	if status == StatusConnected && previousStatus != StatusConnected {
		c.connectionState.ConnectedSince = time.Now()
		c.logger.Info("Connection established, setting ConnectedSince timestamp")
	}
}

// GetConnectionRetries mengembalikan jumlah percobaan koneksi
// Returns: jumlah percobaan koneksi
func (c *Client) GetConnectionRetries() int {
	return c.connectionState.ConnectionRetries
}

// GetWhatsmeowClient mengembalikan referensi ke client whatsmeow
// Returns: pointer ke whatsmeow.Client
func (c *Client) GetWhatsmeowClient() *whatsmeow.Client {
	return c.waClient
}

// RegisterCallback mendaftarkan callback untuk event tertentu
// Parameters:
// - eventName: nama event yang akan ditangani
// - callback: fungsi yang akan dipanggil saat event terjadi
func (c *Client) RegisterCallback(eventName string, callback func(interface{})) {
	c.callbackHandlers[eventName] = callback
}

// UpdateLastActivity memperbarui timestamp aktivitas terakhir
func (c *Client) UpdateLastActivity() {
	c.connectionState.LastActivity = time.Now()
}

// GetSelfID mengembalikan JID dari perangkat sendiri
// Returns: JID perangkat sendiri atau nil jika tidak tersedia
func (c *Client) GetSelfID() *types.JID {
	if c.waClient == nil || !c.waClient.IsLoggedIn() {
		return nil
	}

	return c.waClient.Store.ID
}

// createTimeoutContext membuat context dengan timeout standar
// Parameters:
// - duration: durasi timeout (0 untuk menggunakan default 10 detik)
// Returns:
// - context.Context dengan timeout
// - fungsi cancel yang harus dipanggil dengan defer
func (c *Client) createTimeoutContext(duration time.Duration) (context.Context, context.CancelFunc) {
	if duration <= 0 {
		duration = 10 * time.Second // Default timeout
	}
	return context.WithTimeout(context.Background(), duration)
}
