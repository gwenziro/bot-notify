package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
)

// Connect menginisialisasi klien WhatsApp dan mencoba terhubung
func (c *Client) Connect() error {
	c.logger.Info("Mencoba menghubungkan ke WhatsApp")
	c.connectionState.Status = StatusConnecting

	// Dapatkan device store dari session
	deviceStore, err := c.SessionManager.GetDevice()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan device: %w", err)
	}

	// Buat klien WhatsApp
	waLogger := NewWhatsmeowLogger(c.logger)
	client := whatsmeow.NewClient(deviceStore, waLogger)
	c.waClient = client

	// Daftarkan event handler
	c.eventHandler = c.registerEventHandler()

	// Periksa apakah sudah ada sesi yang disimpan
	if client.Store.ID == nil {
		// Tidak ada sesi, perlu QR code untuk login
		c.logger.Info("Tidak ada sesi WhatsApp yang tersimpan, memulai login dengan QR code")
		qrChan, err := client.GetQRChannel(c.ctx)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan QR channel: %w", err)
		}

		c.logger.Info("Silakan pindai QR code yang akan muncul di terminal...")

		// Mulai koneksi
		err = client.Connect()
		if err != nil {
			return fmt.Errorf("gagal memulai koneksi: %w", err)
		}

		// Tunggu QR code atau berhasil login
		select {
		case qrCode := <-qrChan:
			// Tampilkan QR code untuk dipindai
			qrCodeData := qrCode.Code
			c.logger.Info("QR Code diterima, scan untuk login:")
			c.logger.Info(fmt.Sprintf("QR Code: %s", qrCodeData))

			// Jalankan callback QR Code jika ada
			if callback, ok := c.callbackHandlers["QRCode"]; ok {
				callback(qrCodeData)
			}

			// Tunggu hingga terhubung atau timeout
			for {
				if client.IsConnected() {
					break
				}

				select {
				case <-c.ctx.Done():
					return errors.New("konteks dibatalkan saat menunggu koneksi")
				case <-time.After(time.Second):
					// Cek tiap detik
				}
			}
		case <-c.ctx.Done():
			return errors.New("konteks dibatalkan saat menunggu QR code")
		}
	} else {
		// Sudah ada sesi, coba connect
		c.logger.Info("Sesi WhatsApp ditemukan, mencoba connect")
		err = client.Connect()
		if err != nil {
			return fmt.Errorf("gagal terhubung dengan sesi yang ada: %w", err)
		}
	}

	return nil
}

// Disconnect menutup koneksi WhatsApp dengan bersih
func (c *Client) Disconnect() {
	if c.waClient == nil {
		return
	}

	c.logger.Info("Menutup koneksi WhatsApp")

	// Perbarui status koneksi SEBELUM memanggil disconnect
	c.connectionState.Status = StatusDisconnected
	c.connectionState.IsConnected = false

	// PENTING: Reset ConnectedSince saat disconnect
	c.connectionState.ConnectedSince = time.Time{} // Set ke zero time

	c.connectionState.Timestamp = time.Now()

	// Reset counter pesan saat disconnect
	c.ResetMessageCount()

	// Tutup koneksi aktual
	c.waClient.Disconnect()
}

// handleConnectedEvent menangani event Connected
func (c *Client) handleConnectedEvent() {
	// Dapatkan dan log info perangkat yang terhubung
	contactName := "Unknown"

	if c.waClient != nil && c.waClient.Store != nil {
		// Coba mendapatkan nomor dan nama
		jid := "Unknown"
		if c.waClient.Store.ID != nil {
			jid = c.waClient.Store.ID.String()
		}

		pushName := c.waClient.Store.PushName
		contactName = pushName
		if contactName == "" {
			contactName = jid
		}

		// Cek detail autentikasi
		c.logger.Debug("Store authentication details", utils.Fields{
			"push_name":    pushName,
			"jid":          jid,
			"is_logged_in": c.waClient.IsLoggedIn(),
		})
	}

	c.logger.Info("Terhubung ke WhatsApp", utils.Fields{
		"contact_name":     contactName,
		"client_connected": c.waClient != nil && c.waClient.IsConnected(),
		"client_logged_in": c.waClient != nil && c.waClient.IsLoggedIn(),
	})

	// Update status koneksi
	previousConnected := c.connectionState.IsConnected
	c.connectionState.Status = StatusConnected
	c.connectionState.IsConnected = true

	// Set ConnectedSince HANYA jika sebelumnya tidak terhubung
	// Ini menjamin nilai hanya diperbarui saat pertama kali terhubung
	if !previousConnected {
		c.connectionState.ConnectedSince = time.Now()
		c.logger.Info("Waktu koneksi awal dicatat", utils.Fields{
			"connected_since": c.connectionState.ConnectedSince.Format(time.RFC3339),
			"last_activity":   c.connectionState.LastActivity.Format(time.RFC3339),
		})
	} else {
		c.logger.Debug("Tetap mempertahankan waktu koneksi awal", utils.Fields{
			"connected_since": c.connectionState.ConnectedSince.Format(time.RFC3339),
			"last_activity":   c.connectionState.LastActivity.Format(time.RFC3339),
		})
	}

	// Selalu perbarui LastActivity
	c.UpdateLastActivity()
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

// CheckConnection memeriksa apakah koneksi masih valid dan mencoba koneksi ulang jika perlu
func (c *Client) CheckConnection() (bool, error) {
	if c.waClient == nil {
		c.logger.Warn("WhatsApp client is nil, attempting to reconnect")
		return false, c.Connect()
	}

	// Periksa apakah client menganggap dirinya terhubung
	clientConnected := c.waClient.IsConnected()

	// Jika client menganggap terhubung, tapi state kita mengatakan tidak, perbarui state
	if clientConnected && !c.connectionState.IsConnected {
		c.logger.Info("Client terhubung tetapi state tidak terhubung, memperbaiki state")
		c.SetConnectionState(StatusConnected, true)
		return true, nil
	}

	// Jika client menganggap tidak terhubung, tapi state kita mengatakan ya, coba koneksi ulang
	if !clientConnected && c.connectionState.IsConnected {
		c.logger.Warn("Terdeteksi koneksi terputus saat state terhubung, mencoba koneksi ulang")
		c.SetConnectionState(StatusDisconnected, false)

		// Mencoba koneksi ulang
		err := c.Connect()
		if err != nil {
			c.logger.WithError(err).Error("Gagal melakukan koneksi ulang otomatis")
			return false, err
		}

		c.logger.Info("Berhasil melakukan koneksi ulang otomatis")
		return true, nil
	}

	return clientConnected, nil
}

// PingConnection mengirim ping untuk memastikan koneksi masih hidup
func (c *Client) PingConnection() bool {
	if c.waClient == nil {
		return false
	}

	// Coba operasi sederhana yang menggunakan WebSocket
	isConnected := c.waClient.IsLoggedIn() && c.waClient.IsConnected()

	// Jika tidak terhubung, perbarui status
	if !isConnected && c.connectionState.IsConnected {
		c.logger.Warn("Koneksi terputus terdeteksi selama ping")
		c.SetConnectionState(StatusDisconnected, false)
	}

	return isConnected
}
