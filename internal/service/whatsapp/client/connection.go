package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/service/whatsapp"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
)

// Connect menginisialisasi klien WhatsApp dan mencoba terhubung
// Returns: error jika gagal menghubungkan
func (c *Client) Connect() error {
	c.logger.Info("Mencoba menghubungkan ke WhatsApp")
	c.SetConnectionState(StatusConnecting, false, c.connectionState.ConnectionRetries)

	// Dapatkan device store dari session
	deviceStore, err := c.SessionManager.GetDevice()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan device: %w", err)
	}

	// Buat klien WhatsApp
	waLogger := NewWhatsmeowLogger(c.logger)
	client := whatsmeow.NewClient(deviceStore, waLogger)
	c.waClient = client

	// Periksa apakah sudah ada sesi yang disimpan
	if client.Store.ID == nil {
		// Tidak ada sesi, perlu QR code untuk login
		return c.handleNewSession()
	} else {
		// Sudah ada sesi, coba connect
		return c.handleExistingSession()
	}
}

// handleNewSession menangani koneksi untuk sesi baru dengan QR code
func (c *Client) handleNewSession() error {
	c.logger.Info("Tidak ada sesi WhatsApp yang tersimpan, memulai login dengan QR code")

	qrChan, err := c.waClient.GetQRChannel(c.ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan QR channel: %w", err)
	}

	c.logger.Info("Silakan pindai QR code yang akan muncul...")

	// Mulai koneksi
	err = c.waClient.Connect()
	if err != nil {
		return fmt.Errorf("gagal memulai koneksi: %w", err)
	}

	// Tunggu QR code atau berhasil login dengan timeout
	connectionTimeout := 60 * time.Second
	timeoutCtx, cancel := context.WithTimeout(c.ctx, connectionTimeout)
	defer cancel()

	select {
	case qrCode := <-qrChan:
		// Tampilkan QR code untuk dipindai
		qrCodeData := qrCode.Code
		c.logger.Info("QR Code diterima, scan untuk login")

		// Jalankan callback QR Code jika ada
		if callback, ok := c.callbackHandlers["qr_code"]; ok {
			callback(qrCodeData)
		}

		// Tunggu hingga terhubung atau timeout
		return c.waitForConnection(timeoutCtx)

	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return whatsapp.ErrConnectionTimeout
		}
		return whatsapp.ErrContextCanceled
	}
}

// handleExistingSession menangani koneksi untuk sesi yang sudah ada
func (c *Client) handleExistingSession() error {
	c.logger.Info("Sesi WhatsApp ditemukan, mencoba connect")

	// Mulai koneksi
	err := c.waClient.Connect()
	if err != nil {
		return fmt.Errorf("gagal terhubung dengan sesi yang ada: %w", err)
	}

	// Tunggu hingga terhubung dengan timeout
	connectionTimeout := 30 * time.Second
	timeoutCtx, cancel := context.WithTimeout(c.ctx, connectionTimeout)
	defer cancel()

	return c.waitForConnection(timeoutCtx)
}

// waitForConnection menunggu hingga koneksi terbentuk atau timeout
func (c *Client) waitForConnection(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.waClient != nil && c.waClient.IsConnected() {
				c.logger.Info("WhatsApp berhasil terhubung!")
				return nil
			}
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return whatsapp.ErrConnectionTimeout
			}
			return whatsapp.ErrContextCanceled
		}
	}
}

// Disconnect menutup koneksi WhatsApp dengan bersih
func (c *Client) Disconnect() {
	if c.waClient == nil {
		return
	}

	c.logger.Info("Menutup koneksi WhatsApp")

	// Perbarui status koneksi SEBELUM memanggil disconnect
	c.SetConnectionState(StatusDisconnected, false, 0)

	// Tutup koneksi aktual
	c.waClient.Disconnect()
}

// AttemptReconnect mencoba reconnect dengan exponential backoff
// Parameters:
// - reason: alasan melakukan reconnect (untuk logging)
func (c *Client) AttemptReconnect(reason string) {
	c.reconnectLock.Lock()
	defer c.reconnectLock.Unlock()

	// Pastikan tidak ada percobaan reconnect yang sedang berlangsung
	if c.retryTimer != nil {
		c.retryTimer.Stop()
	}

	// Batas maksimum percobaan
	if c.connectionState.ConnectionRetries >= c.config.MaxRetry {
		c.logger.WithFields(utils.Fields{
			"max_retries": c.config.MaxRetry,
			"reason":      reason,
		}).Error("Mencapai batas maksimum percobaan reconnect")
		return
	}

	// Tingkatkan counter percobaan
	newRetryCount := c.connectionState.ConnectionRetries + 1
	c.SetConnectionState(StatusConnecting, false, newRetryCount)

	// Hitung waktu delay dengan exponential backoff
	delay := time.Duration(1<<uint(newRetryCount-1)) * time.Second
	if delay > c.config.RetryDelay {
		delay = c.config.RetryDelay
	}

	c.logger.WithFields(utils.Fields{
		"delay":   delay,
		"attempt": newRetryCount,
		"reason":  reason,
	}).Info("Mencoba reconnect")

	// Set timer untuk reconnect
	c.retryTimer = time.AfterFunc(delay, func() {
		// Bersihkan resource lama jika ada
		if c.waClient != nil {
			c.waClient.Disconnect()
		}

		// Coba connect ulang
		err := c.Connect()
		if err != nil {
			c.logger.WithFields(utils.Fields{
				"error":   err,
				"attempt": newRetryCount,
			}).Error("Gagal reconnect")

			// Coba lagi dengan AttemptReconnect
			c.AttemptReconnect("reconnect_failed")
		}
	})
}

// Close menutup semua resource yang digunakan oleh klien
func (c *Client) Close() {
	c.logger.Info("Menutup klien WhatsApp dan membersihkan resource")

	// Batalkan konteks untuk menghentikan operasi yang sedang berlangsung
	c.cancel()

	// Tutup koneksi WhatsApp
	if c.waClient != nil {
		c.waClient.Disconnect()
	}

	// Tutup device store
	if c.deviceStore != nil {
		c.deviceStore.Close()
	}
}

// validateConnection memeriksa apakah klien WhatsApp terhubung
func (c *Client) validateConnection() error {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return whatsapp.ErrWhatsAppNotConnected
	}
	return nil
}

// IsLoggedIn memeriksa apakah pengguna sudah login
// Returns: true jika sudah login
func (c *Client) IsLoggedIn() bool {
	if c.waClient == nil {
		return false
	}

	return c.waClient.Store.ID != nil
}

// GetConnectionInfo mendapatkan informasi lengkap tentang koneksi
// Returns: map dengan informasi koneksi
func (c *Client) GetConnectionInfo() map[string]interface{} {
	state := c.connectionState

	return map[string]interface{}{
		"status":      state.Status,
		"connected":   state.IsConnected,
		"last_active": state.LastActivity,
		"retry_count": state.ConnectionRetries,
		"logged_in":   c.IsLoggedIn(),
	}
}

// GetConnectionStateSafe mengembalikan state koneksi dengan pengecekan null
// Returns:
// - state koneksi
// - error jika gagal
func (c *Client) GetConnectionStateSafe() (ConnectionState, error) {
	// Cek untuk mencegah nil dereference
	if c == nil {
		return ConnectionState{
			Status:            StatusDisconnected,
			IsConnected:       false,
			ConnectionRetries: 0,
			LastActivity:      time.Now(),
			Timestamp:         time.Now(),
		}, fmt.Errorf("client adalah nil")
	}

	// Deep copy untuk mencegah race condition
	state := ConnectionState{
		Status:            c.connectionState.Status,
		IsConnected:       c.connectionState.IsConnected,
		ConnectionRetries: c.connectionState.ConnectionRetries,
		LastActivity:      c.connectionState.LastActivity,
		Timestamp:         c.connectionState.Timestamp,
		ConnectedSince:    c.connectionState.ConnectedSince,
	}

	return state, nil
}
