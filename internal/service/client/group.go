package client

import (
	"errors"
	"fmt"

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

	// Konversi ID ke JID menggunakan utils
	jid := utils.ParseGroupID(groupID)

	// Ambil info grup
	group, err := c.waClient.GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info grup %s: %w", groupID, err)
	}

	return group, nil
}

// GetGroupParticipants mendapatkan daftar anggota grup dengan informasi lengkap
func (c *Client) GetGroupParticipants(groupJID types.JID) ([]types.GroupParticipant, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return nil, fmt.Errorf("klien WhatsApp belum terhubung")
	}

	// Dapatkan informasi grup termasuk anggota
	group, err := c.waClient.GetGroupInfo(groupJID)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info grup: %w", err)
	}

	c.logger.Debug("Retrieved group participants", utils.Fields{
		"group_id":         groupJID.String(),
		"participants":     len(group.Participants),
		"has_participants": group.Participants != nil,
	})

	return group.Participants, nil
}

// IsGroupAdmin memeriksa apakah pengguna adalah admin grup
func (c *Client) IsGroupAdmin(groupJID types.JID) (bool, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return false, fmt.Errorf("klien WhatsApp belum terhubung")
	}

	// Dapatkan ID sendiri
	selfID := c.GetSelfID()
	if selfID == nil {
		return false, fmt.Errorf("ID akun tidak tersedia")
	}

	// Dapatkan anggota grup
	participants, err := c.GetGroupParticipants(groupJID)
	if err != nil {
		return false, err
	}

	// Cari diri sendiri dalam daftar anggota
	for _, participant := range participants {
		// Bandingkan JID tanpa bagian device menggunakan utils
		selfJIDStr := utils.NormalizeJID(selfID.String())
		participantJIDStr := utils.NormalizeJID(participant.JID.String())

		if selfJIDStr == participantJIDStr {
			return participant.IsAdmin, nil
		}
	}

	return false, fmt.Errorf("pengguna tidak ditemukan dalam grup")
}

// GetEnrichedParticipants mendapatkan daftar peserta grup dengan informasi tambahan
func (c *Client) GetEnrichedParticipants(groupJID types.JID) ([]types.GroupParticipant, error) {
	// Dapatkan partisipan dasar
	participants, err := c.GetGroupParticipants(groupJID)
	if err != nil {
		return nil, err
	}

	// Tidak perlu enrichment jika tidak ada partisipan
	if len(participants) == 0 {
		return participants, nil
	}

	// Enrich each participant with contact information
	for i := range participants {
		// Jika DisplayName kosong, coba isi dari kontak
		if participants[i].DisplayName == "" {
			contactInfo, err := c.GetContactInfo(participants[i].JID.String())
			if err == nil && contactInfo != nil {
				participants[i].DisplayName = contactInfo.FullName
			}
		}
	}

	return participants, nil
}
