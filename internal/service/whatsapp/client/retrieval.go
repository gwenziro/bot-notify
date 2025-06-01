package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/service/whatsapp"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// GetContactPictureURL mendapatkan URL foto profil kontak atau diri sendiri
// Parameters:
// - jid: JID kontak (nil untuk foto diri sendiri)
// Returns:
// - URL foto profil
// - error jika gagal
func (c *Client) GetContactPictureURL(jid *types.JID) (string, error) {
	if err := c.validateConnection(); err != nil {
		return "", err
	}

	if !c.waClient.IsLoggedIn() {
		return "", errors.New("tidak dalam keadaan login")
	}

	// Jika JID tidak disediakan, gunakan ID diri sendiri
	targetJID := jid
	if targetJID == nil {
		selfID := c.waClient.Store.ID
		if selfID == nil {
			return "", whatsapp.ErrSelfIDNotAvailable
		}
		targetJID = selfID
	}

	// Buat timeout context
	_, cancel := c.createTimeoutContext(10 * time.Second)
	defer cancel()

	// Gunakan API whatsmeow untuk mendapatkan foto profil
	profilePic, err := c.waClient.GetProfilePictureInfo(targetJID.ToNonAD(), &whatsmeow.GetProfilePictureParams{
		Preview: false,
	})
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan info foto profil: %w", err)
	}

	if profilePic == nil || profilePic.URL == "" {
		return "", nil
	}

	return profilePic.URL, nil
}

// GetContactName mendapatkan nama kontak terbaik dari JID yang tersedia
// Parameters:
// - jid: JID kontak yang dicari
// - defaultName: nama default jika tidak ada nama yang tersedia
// Returns: nama kontak terbaik berdasarkan prioritas
func (c *Client) GetContactName(jid types.JID, defaultName string) string {
	if c.waClient == nil || !c.waClient.IsLoggedIn() {
		return defaultName
	}

	// Cek apakah ini JID diri sendiri
	isSelf := c.waClient.Store.ID != nil && c.waClient.Store.ID.User == jid.User

	// Untuk diri sendiri, prioritaskan pushName dari store
	if isSelf && c.waClient.Store.PushName != "" {
		return c.waClient.Store.PushName
	}

	// Konversi ke non-AD JID jika perlu
	nonAD := jid.ToNonAD()

	// Coba dapatkan dari store kontak dengan timeout
	ctx, cancel := c.createTimeoutContext(3 * time.Second)
	defer cancel()

	contact, err := c.waClient.Store.Contacts.GetContact(ctx, nonAD)
	if err != nil {
		c.logger.Debug("Tidak dapat mendapatkan kontak", utils.Fields{
			"jid": jid.String(),
			"err": err.Error(),
		})
		return defaultName
	}

	// Prioritaskan nama berdasarkan yang tersedia
	if contact.FullName != "" {
		return contact.FullName
	} else if contact.PushName != "" {
		return contact.PushName
	} else if contact.BusinessName != "" {
		return contact.BusinessName
	}

	return defaultName
}
