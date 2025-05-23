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
					c.logger.Info("WhatsApp berhasil terhubung!")
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
	c.waClient.Disconnect()
}

// AttemptReconnect mencoba reconnect dengan exponential backoff
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

	c.connectionState.ConnectionRetries++

	// Hitung waktu delay dengan exponential backoff
	delay := time.Duration(1<<uint(c.connectionState.ConnectionRetries-1)) * time.Second
	if delay > c.config.RetryDelay {
		delay = c.config.RetryDelay
	}

	c.logger.WithFields(utils.Fields{
		"delay":   delay,
		"attempt": c.connectionState.ConnectionRetries,
		"reason":  reason,
	}).Info("Mencoba reconnect")

	c.connectionState.Status = StatusConnecting

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
				"attempt": c.connectionState.ConnectionRetries,
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
