package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// GetGroups mengembalikan daftar grup yang tersedia
func (c *Client) GetGroups() ([]*types.GroupInfo, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.Info("Mengambil daftar grup")
	c.UpdateLastActivity()

	groups, err := c.waClient.GetJoinedGroups()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan daftar grup: %w", err)
	}

	return groups, nil
}

// GetGroupByID mencari grup berdasarkan ID
func (c *Client) GetGroupByID(groupID string) (*types.GroupInfo, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, errors.New("klien WhatsApp belum terhubung")
	}

	// Konversi ID ke JID
	jid := ParseGroupID(groupID)

	// Ambil info grup
	group, err := c.waClient.GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info grup %s: %w", groupID, err)
	}

	return group, nil
}

// GetContactInfo mendapatkan informasi kontak
func (c *Client) GetContactInfo(phoneNumber string) (*types.ContactInfo, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, errors.New("klien WhatsApp belum terhubung")
	}

	// Konversi nomor telepon ke JID
	jid := ParsePhoneNumber(phoneNumber)
	// Ambil info kontak
	c.logger.WithField("phone", phoneNumber).Debug("Mengambil informasi kontak")
	c.UpdateLastActivity()

	ctx := context.Background()
	contact, err := c.waClient.Store.Contacts.GetContact(ctx, jid)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info kontak %s: %w", phoneNumber, err)
	}

	return &contact, nil
}

// IsLoggedIn memeriksa apakah pengguna sudah login
func (c *Client) IsLoggedIn() bool {
	if c.waClient == nil {
		return false
	}

	return c.waClient.Store.ID != nil
}

// GetConnectionInfo mendapatkan informasi lengkap tentang koneksi
func (c *Client) GetConnectionInfo() map[string]interface{} {
	state := c.connectionState

	return map[string]interface{}{
		"status":      state.Status,
		"connected":   state.IsConnected,
		"last_active": state.LastActivity,
		"retry_count": state.ConnectionRetries,
		"logged_in":   c.IsLoggedIn(),
		"device_info": c.GetDeviceInfo(),
	}
}

// GetConnectionStateSafe mengembalikan state koneksi dengan pengecekan null
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
	}

	return state, nil
}

// GetDeviceInfo mengembalikan informasi device yang digunakan
func (c *Client) GetDeviceInfo() map[string]interface{} {
	if c == nil {
		return map[string]interface{}{
			"logged_in": false,
			"error":     "client is nil",
		}
	}

	if c.waClient == nil {
		return map[string]interface{}{
			"logged_in": false,
			"error":     "whatsmeow client is nil",
		}
	}

	if c.waClient.Store == nil {
		return map[string]interface{}{
			"logged_in": false,
			"error":     "whatsmeow store is nil",
		}
	}

	if c.waClient.Store.ID == nil {
		return map[string]interface{}{
			"logged_in": false,
			"error":     "not logged in, no JID available",
		}
	}

	// Log semua informasi yang tersedia di store
	c.logger.Debug("WhatsApp store info", utils.Fields{
		"store_id":     c.waClient.Store.ID.String(),
		"push_name":    c.waClient.Store.PushName,
		"has_contacts": c.waClient.Store.Contacts != nil,
	})

	// Pastikan push_name diambil dengan benar
	pushName := c.waClient.Store.PushName
	c.logger.Debug("Push name value", utils.Fields{"push_name": pushName})

	// Jika pushName kosong, coba cari dari sumber lain
	if pushName == "" {
		c.logger.Debug("Push name is empty, trying to get it from account info")

		// Coba dapatkan info tambahan
		if c.waClient.IsConnected() && c.waClient.IsLoggedIn() {
			_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
		}

	}

	// Format JID untuk tampilan yang lebih mudah dibaca
	jid := c.waClient.Store.ID.String()
	formattedNumber := FormatWhatsAppNumber(jid)

	deviceInfo := map[string]interface{}{
		"id":            jid,
		"formatted_jid": formattedNumber,
		"logged_in":     true,
		"push_name":     pushName,
	}

	// Log informasi device untuk debugging
	c.logger.Debug("Device info", utils.Fields{
		"id":        jid,
		"number":    formattedNumber,
		"push_name": pushName,
	})

	return deviceInfo
}
