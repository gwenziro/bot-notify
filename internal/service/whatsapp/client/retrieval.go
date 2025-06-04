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

	// Konversi ID ke JID menggunakan utils
	jid := utils.ParseGroupID(groupID)

	// Ambil info grup
	group, err := c.waClient.GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info grup %s: %w", groupID, err)
	}

	return group, nil
}

// GetContactInfo mendapatkan informasi kontak berdasarkan nomor telepon atau JID
func (c *Client) GetContactInfo(identifier string) (*types.ContactInfo, error) {
	// Validasi koneksi
	if err := c.validateConnection(); err != nil {
		return nil, err
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

// validateConnection adalah helper untuk memvalidasi koneksi
func (c *Client) validateConnection() error {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return errors.New("klien WhatsApp belum terhubung")
	}

	if c.waClient.Store == nil || c.waClient.Store.Contacts == nil {
		return errors.New("penyimpanan kontak WhatsApp tidak tersedia")
	}

	return nil
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
		"status":          state.Status,
		"connected":       state.IsConnected,
		"last_active":     state.LastActivity,
		"connected_since": state.ConnectedSince,
		"logged_in":       c.IsLoggedIn(),
		"device_info":     c.GetDeviceInfo(),
	}
}

// GetConnectionStateSafe mengembalikan state koneksi dengan pengecekan null
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
	if c.waClient == nil || !c.waClient.IsLoggedIn() {
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

// GetDeviceInfo mengembalikan informasi tentang perangkat WhatsApp yang terhubung
func (c *Client) GetDeviceInfo() map[string]interface{} {
	// Validasi dasar
	if c == nil || c.waClient == nil || c.waClient.Store == nil || c.waClient.Store.ID == nil {
		return map[string]interface{}{
			"logged_in": false,
			"error":     "tidak ada sesi aktif",
		}
	}

	// Dapatkan informasi dasar yang pasti tersedia
	jid := c.waClient.Store.ID.String()
	formattedNumber := utils.FormatWhatsAppNumber(jid)
	pushName := c.waClient.Store.PushName

	// Dapatkan URL foto profil jika tersedia
	pictureURL := ""
	if c.waClient.IsLoggedIn() {
		if pic, err := c.GetProfilePictureURL(); err == nil && pic != "" {
			pictureURL = pic
		}
	}

	// Siapkan hasil
	result := map[string]interface{}{
		"id":            jid,
		"formatted_jid": formattedNumber,
		"logged_in":     true,
		"push_name":     pushName,
		"picture_url":   pictureURL,
	}

	c.logger.Debug("Device info retrieved", utils.Fields{
		"jid":         jid,
		"number":      formattedNumber,
		"push_name":   pushName,
		"has_picture": pictureURL != "",
	})

	return result
}

// GetContactNameByJID mendapatkan nama kontak dari JID jika tersedia
func (c *Client) GetContactNameByJID(jid types.JID) string {
	if c.waClient == nil || !c.waClient.IsLoggedIn() {
		return ""
	}

	// Konversi ke non-AD JID jika perlu
	nonAD := jid.ToNonAD()

	// Coba dapatkan dari store kontak
	ctx := context.Background()
	contact, err := c.waClient.Store.Contacts.GetContact(ctx, nonAD)
	if err != nil {
		c.logger.Debug("Tidak dapat mendapatkan kontak", utils.Fields{
			"jid": jid.String(),
			"err": err.Error(),
		})
		return ""
	}

	// Prioritaskan nama berdasarkan yang tersedia
	if contact.FullName != "" {
		return contact.FullName
	} else if contact.PushName != "" {
		return contact.PushName
	} else if contact.BusinessName != "" {
		return contact.BusinessName
	}

	return ""
}
