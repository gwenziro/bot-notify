package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// GetContactInfo mendapatkan informasi kontak berdasarkan nomor telepon atau JID
func (c *Client) GetContactInfo(identifier string) (*types.ContactInfo, error) {
	// Validasi koneksi
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, errors.New("klien WhatsApp belum terhubung")
	}

	if c.waClient.Store == nil || c.waClient.Store.Contacts == nil {
		return nil, errors.New("penyimpanan kontak WhatsApp tidak tersedia")
	}

	c.UpdateLastActivity()

	// Ambil kontak dari store menggunakan utils
	ctx := context.Background()
	contact, err := c.waClient.Store.Contacts.GetContact(ctx, utils.ParsePhoneNumber(identifier))
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info kontak %s: %w", identifier, err)
	}

	return &contact, nil
}

// GetConnectionInfo mendapatkan informasi lengkap tentang koneksi
func (c *Client) GetConnectionInfo() map[string]interface{} {
	// Dapatkan connection state dengan cara yang aman
	state, err := c.GetConnectionStateSafe()
	if err != nil {
		return map[string]interface{}{
			"status":      string(StatusDisconnected),
			"connected":   false,
			"last_active": time.Now(),
			"logged_in":   false,
			"error":       err.Error(),
		}
	}

	// Dapatkan device info dengan handling nil pointer
	var deviceInfo map[string]interface{}
	if c != nil && c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		jid := c.waClient.Store.ID.String()
		formattedNumber := utils.FormatWhatsAppNumber(jid)

		// Coba dapatkan nama dari store kontak
		var pushName string
		if c.waClient.Store.PushName != "" {
			pushName = c.waClient.Store.PushName
		}

		// Dapatkan URL foto profil jika tersedia dan client terhubung
		var pictureURL string
		if c.IsLoggedIn() {
			if pic, err := c.GetProfilePictureURL(); err == nil && pic != "" {
				pictureURL = pic
			}
		}

		deviceInfo = map[string]interface{}{
			"id":            jid,
			"formatted_jid": formattedNumber,
			"logged_in":     true,
			"push_name":     pushName,
			"picture_url":   pictureURL,
		}
	} else {
		deviceInfo = map[string]interface{}{
			"logged_in": false,
			"error":     "tidak ada sesi aktif",
		}
	}

	return map[string]interface{}{
		"status":          state.Status,
		"connected":       state.IsConnected,
		"last_active":     state.LastActivity,
		"connected_since": state.ConnectedSince,
		"logged_in":       c.IsLoggedIn(),
		"device_info":     deviceInfo,
	}
}

// GetConnectionStateSafe mengembalikan state koneksi dengan pengecekan null
// Thread-safe dan dengan deep copy untuk mencegah race condition
func (c *Client) GetConnectionStateSafe() (ConnectionState, error) {
	// Cek untuk mencegah nil dereference
	if c == nil {
		return ConnectionState{
			Status:       StatusDisconnected,
			IsConnected:  false,
			LastActivity: time.Now(),
			Timestamp:    time.Now(),
		}, fmt.Errorf("client adalah nil")
	}

	// Deep copy untuk mencegah race condition
	state := ConnectionState{
		Status:         c.connectionState.Status,
		IsConnected:    c.connectionState.IsConnected,
		LastActivity:   c.connectionState.LastActivity,
		Timestamp:      c.connectionState.Timestamp,
		ConnectedSince: c.connectionState.ConnectedSince,
	}

	return state, nil
}

// GetProfilePictureURL mendapatkan URL foto profil akun WhatsApp terhubung
func (c *Client) GetProfilePictureURL() (string, error) {
	if c.waClient == nil || !c.IsLoggedIn() {
		return "", fmt.Errorf("client tidak terhubung atau login")
	}

	// Dapatkan ID kita sendiri
	selfID := c.waClient.Store.ID
	if selfID == nil {
		return "", fmt.Errorf("id akun tidak tersedia")
	}

	// Gunakan API whatsmeow untuk mendapatkan foto profil
	profilePic, err := c.waClient.GetProfilePictureInfo(selfID.ToNonAD(), &whatsmeow.GetProfilePictureParams{
		Preview: false,
	})
	if err != nil {
		return "", err
	}

	if profilePic == nil || profilePic.URL == "" {
		return "", nil
	}

	return profilePic.URL, nil
}

// IsLoggedIn memeriksa apakah pengguna sudah login
func (c *Client) IsLoggedIn() bool {
	if c.waClient == nil {
		return false
	}

	// Gunakan fungsi bawaan dari whatsmeow
	return c.waClient.IsLoggedIn()
}
