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

	_ "modernc.org/sqlite" // Pure Go SQLite driver
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
	Status            ClientStatus `json:"status"`
	IsConnected       bool         `json:"is_connected"`
	ConnectionRetries int          `json:"connection_retries"`
	LastActivity      time.Time    `json:"last_activity"`
	Timestamp         time.Time    `json:"timestamp"`
}

// EventHandlerFunc adalah tipe fungsi untuk menangani event WhatsApp
type EventHandlerFunc func(interface{})

// Client adalah wrapper untuk klien WhatsApp
type Client struct {
	waClient        *whatsmeow.Client
	eventHandler    uint32
	deviceStore     *sqlstore.Container
	connectionState ConnectionState
	config          *config.WhatsAppConfig
	logger          utils.LogrusEntry
	qrChan          chan string
	SessionManager  *session.Manager

	callbackHandlers map[string]func(interface{})
	reconnectLock    sync.Mutex
	retryTimer       *time.Timer
	ctx              context.Context
	cancel           context.CancelFunc
}

// GetConnectionState mengembalikan state koneksi saat ini
func (c *Client) GetConnectionState() ConnectionState {
	return c.connectionState
}

// SetConnectionState mengatur status koneksi saat ini
func (c *Client) SetConnectionState(status ClientStatus, isConnected bool, retries int) {
	c.connectionState.Status = status
	c.connectionState.IsConnected = isConnected
	c.connectionState.ConnectionRetries = retries
	c.connectionState.Timestamp = time.Now()
}

// GetConnectionRetries mengembalikan jumlah percobaan koneksi
func (c *Client) GetConnectionRetries() int {
	return c.connectionState.ConnectionRetries
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

// UpdateLastActivity memperbarui timestamp aktivitas terakhir
func (c *Client) UpdateLastActivity() {
	c.connectionState.LastActivity = time.Now()
}

// NewClient membuat instance baru dari klien WhatsApp
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
