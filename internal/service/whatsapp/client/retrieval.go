package client

import (
	"context"
	"errors"
	"fmt"

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

// GetDeviceInfo mengembalikan informasi device yang digunakan
func (c *Client) GetDeviceInfo() map[string]interface{} {
	if c.waClient == nil || c.waClient.Store.ID == nil {
		return map[string]interface{}{
			"logged_in": false,
		}
	}

	return map[string]interface{}{
		"id":        c.waClient.Store.ID.String(),
		"logged_in": true,
		"push_name": c.waClient.Store.PushName,
	}
}
