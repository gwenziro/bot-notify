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

	params := &whatsmeow.GetProfilePictureParams{
		Preview: false,
	}

	// Gunakan API whatsmeow untuk mendapatkan foto profil
	profilePic, err := c.waClient.GetProfilePictureInfo(selfID.ToNonAD(), params)
	if err != nil {
		return "", err
	}

	if profilePic == nil || profilePic.URL == "" {
		return "", nil
	}

	return profilePic.URL, nil
}

// GetOwnContactInfo mendapatkan informasi kontak dari akun WhatsApp yang terhubung saat ini
func (c *Client) GetOwnContactInfo() (*types.ContactInfo, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, errors.New("klien WhatsApp belum terhubung")
	}

	// Dapatkan ID sendiri
	selfID := c.waClient.Store.ID
	if selfID == nil {
		return nil, errors.New("ID akun tidak tersedia")
	}

	// Menggunakan fungsi GetContactInfo yang sudah ada dengan ID sendiri
	selfJID := selfID.ToNonAD()
	selfNumber := FormatWhatsAppNumber(selfJID.String())

	c.logger.Debug("Mengambil informasi kontak sendiri", utils.Fields{
		"jid":    selfJID.String(),
		"number": selfNumber,
	})

	// Ambil kontak menggunakan konteks
	ctx := context.Background()
	contact, err := c.waClient.Store.Contacts.GetContact(ctx, selfJID)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info kontak sendiri: %w", err)
	}

	return &contact, nil
}

// GetDeviceInfo mengembalikan informasi lengkap tentang perangkat yang digunakan
func (c *Client) GetDeviceInfo() map[string]interface{} {
	// Validasi dasar
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

	// Dapatkan informasi dasar
	jid := c.waClient.Store.ID.String()
	formattedNumber := FormatWhatsAppNumber(jid)
	pushName := c.waClient.Store.PushName

	// Dapatkan nama dan informasi kontak tambahan dari GetOwnContactInfo
	if pushName == "" {
		c.logger.Debug("Push name is empty, trying to get it from contact info")

		// Gunakan GetOwnContactInfo untuk mendapatkan informasi kontak
		contactInfo, err := c.GetOwnContactInfo()
		if err == nil && contactInfo != nil && contactInfo.PushName != "" {
			pushName = contactInfo.PushName
			c.logger.Debug("Found push name from own contact info", utils.Fields{"push_name": pushName})
		} else {
			c.logger.Debug("Failed to get contact info", utils.Fields{"error": err})
		}
	}

	// Dapatkan status - catatan: ContactInfo tidak memiliki field Status
	status := "" // Tidak dapat mengakses status kontak

	// Informasi platform - Store.Platform mungkin bukan pointer di versi whatsmeow
	platform := "WhatsApp Web" // Default platform

	// Coba dapatkan informasi platform dari sumber lain
	if c.waClient.Store != nil && c.waClient.Store.PushName != "" {
		// Mungkin kita bisa mendapatkan informasi dari properti lain
		// Untuk saat ini, gunakan default
	}

	// Informasi versi - Store.ClientVersion tidak tersedia di whatsmeow
	version := ""

	// Coba dapatkan URL foto profil
	pictureURL := ""
	if c.waClient.IsLoggedIn() {
		params := &whatsmeow.GetProfilePictureParams{
			Preview: false,
		}
		if profilePic, err := c.waClient.GetProfilePictureInfo(c.waClient.Store.ID.ToNonAD(), params); err == nil && profilePic != nil && profilePic.URL != "" {
			pictureURL = profilePic.URL
			c.logger.Debug("Found profile picture URL", utils.Fields{"url": pictureURL})
		} else {
			c.logger.Debug("Failed to get profile picture", utils.Fields{"error": err})
		}
	}

	// Siapkan hasil yang lengkap
	deviceInfo := map[string]interface{}{
		"id":            jid,
		"formatted_jid": formattedNumber,
		"logged_in":     true,
		"push_name":     pushName,
		"status":        status,
		"platform":      platform,
		"picture_url":   pictureURL,
	}

	// Tambahkan informasi opsional jika tersedia
	if version != "" {
		deviceInfo["version"] = version
	}

	// Buat field device yang menggabungkan platform dan versi
	deviceType := platform
	if version != "" {
		deviceType = fmt.Sprintf("%s %s", platform, version)
	}
	deviceInfo["device"] = deviceType

	// Log informasi device untuk debugging
	c.logger.Debug("Complete device info retrieved", utils.Fields{
		"id":          jid,
		"number":      formattedNumber,
		"push_name":   pushName,
		"platform":    platform,
		"version":     version,
		"has_picture": pictureURL != "",
	})

	return deviceInfo
}
