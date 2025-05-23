package client

import (
	"fmt"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types/events"
)

// registerEventHandler mendaftarkan handler untuk event WhatsApp
func (c *Client) registerEventHandler() uint32 {
	if c.waClient == nil {
		c.logger.Error("Whatsmeow client belum diinisialisasi")
		return 0
	}

	// Daftarkan handler untuk berbagai tipe event
	return c.waClient.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.QR:
			c.handleQRCodeEvent(v)
		case *events.Connected:
			c.handleConnectedEvent()
		case *events.Disconnected:
			c.handleDisconnectedEvent(v)
		case *events.LoggedOut:
			c.handleLoggedOutEvent(v)
		case *events.Message:
			c.handleMessageEvent(v)
		}

		// Panggil callback kustom jika ada
		if handler, exists := c.callbackHandlers[fmt.Sprintf("%T", evt)]; exists {
			handler(evt)
		}
	})
}

// handleQRCodeEvent menangani event QR Code
func (c *Client) handleQRCodeEvent(evt *events.QR) {
	// Kode QR diterima dari server WhatsApp
	var qrCodeStr string
	if len(evt.Codes) > 0 { // Versi baru menggunakan array Codes
		qrCodeStr = evt.Codes[0]
	} else { // Tidak ada kode QR yang tersedia
		c.logger.Warn("QR code event diterima tapi tidak ada kode yang tersedia")
		qrCodeStr = ""
	}

	if qrCodeStr == "" {
		return
	}

	// Log QR code information
	c.logger.Info("QR code diterima, siap untuk dipindai")

	// Jalankan callback QR Code jika ada
	if callback, ok := c.callbackHandlers["QRCode"]; ok {
		callback(qrCodeStr)
	}
}

// handleConnectedEvent menangani event Connected
func (c *Client) handleConnectedEvent() {
	c.logger.Info("Terhubung ke WhatsApp")
	c.connectionState.Status = StatusConnected
	c.connectionState.IsConnected = true
	c.connectionState.ConnectionRetries = 0 // Reset retry counter pada koneksi berhasil
	c.UpdateLastActivity()
}

// handleDisconnectedEvent menangani event Disconnected
func (c *Client) handleDisconnectedEvent(_ *events.Disconnected) {
	c.logger.Warn("Terputus dari WhatsApp")
	c.connectionState.Status = StatusDisconnected
	c.connectionState.IsConnected = false
	// Coba reconnect jika disconnected
	go c.AttemptReconnect("disconnected")
}

// handleLoggedOutEvent menangani event LoggedOut
func (c *Client) handleLoggedOutEvent(_ *events.LoggedOut) {
	c.logger.Warn("Logged out dari WhatsApp")
	c.connectionState.Status = StatusLoggedOut
	c.connectionState.IsConnected = false
}

// handleMessageEvent menangani event pesan masuk
func (c *Client) handleMessageEvent(evt *events.Message) {
	c.logger.WithFields(utils.Fields{
		"from":     evt.Info.Sender.String(),
		"chat":     evt.Info.Chat.String(),
		"is_group": evt.Info.IsGroup,
	}).Debug("Pesan diterima")

	// Update aktivitas terakhir ketika menerima pesan
	c.UpdateLastActivity()
}
