package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/service/whatsapp"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// GetGroups mengembalikan daftar grup yang tersedia
// Returns:
// - daftar grup
// - error jika gagal
func (c *Client) GetGroups() ([]*types.GroupInfo, error) {
	if err := c.validateConnection(); err != nil {
		return nil, err
	}

	c.logger.Info("Mengambil daftar grup")
	c.UpdateLastActivity()

	// Tambahkan timeout operasi
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	groups, err := c.waClient.GetJoinedGroups()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan daftar grup: %w", err)
	}

	c.logger.WithFields(utils.Fields{
		"count": len(groups),
	}).Debug("Berhasil mengambil daftar grup")

	return groups, nil
}

// GetGroupByID mencari grup berdasarkan ID
// Parameters:
// - groupID: ID grup yang akan dicari
// Returns:
// - informasi grup
// - error jika gagal
func (c *Client) GetGroupByID(groupID string) (*types.GroupInfo, error) {
	if err := c.validateConnection(); err != nil {
		return nil, err
	}

	// Konversi ID ke JID
	jidStr := utils.FormatGroupID(groupID)
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing JID grup %s: %w", groupID, err)
	}

	// Tambahkan timeout operasi
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ambil info grup
	group, err := c.waClient.GetGroupInfo(jid)
	if err != nil {
		c.logger.WithFields(utils.Fields{
			"group_id": groupID,
			"error":    err.Error(),
		}).Debug("Gagal mendapatkan info grup")
		return nil, fmt.Errorf("gagal mendapatkan info grup %s: %w", groupID, err)
	}

	if group == nil {
		return nil, whatsapp.ErrGroupNotFound
	}

	return group, nil
}

// GetGroupParticipants mendapatkan daftar anggota grup dengan informasi lengkap
// Parameters:
// - groupJID: JID grup
// Returns:
// - daftar anggota grup
// - error jika gagal
func (c *Client) GetGroupParticipants(groupJID types.JID) ([]types.GroupParticipant, error) {
	if err := c.validateConnection(); err != nil {
		return nil, err
	}

	// Dapatkan informasi grup termasuk anggota
	group, err := c.waClient.GetGroupInfo(groupJID)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info grup: %w", err)
	}

	if group == nil {
		return nil, whatsapp.ErrGroupNotFound
	}

	c.logger.Debug("Retrieved group participants", utils.Fields{
		"group_id":         groupJID.String(),
		"participants":     len(group.Participants),
		"has_participants": group.Participants != nil,
	})

	return group.Participants, nil
}

// GetEnrichedParticipants mendapatkan daftar peserta grup dengan informasi tambahan
// Parameters:
// - groupJID: JID grup
// Returns:
// - daftar peserta grup dengan informasi tambahan
// - error jika gagal
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
			// Gunakan langsung fungsi GetContactName
			contactName := c.GetContactName(participants[i].JID, "")
			participants[i].DisplayName = contactName
		}
	}

	return participants, nil
}

// findParticipantByJID mencari partisipan dalam daftar berdasarkan JID
// Parameters:
// - participants: daftar partisipan grup
// - targetJID: JID yang dicari
// Returns:
// - partisipan yang ditemukan (nil jika tidak ada)
// - index partisipan dalam array
// - apakah ditemukan
func (c *Client) findParticipantByJID(participants []types.GroupParticipant, targetJID types.JID) (*types.GroupParticipant, int, bool) {
	targetJIDStr := utils.NormalizeJID(targetJID.String())

	for i, participant := range participants {
		participantJIDStr := utils.NormalizeJID(participant.JID.String())
		if participantJIDStr == targetJIDStr {
			return &participants[i], i, true
		}
	}

	return nil, -1, false
}

// IsGroupAdmin memeriksa apakah pengguna adalah admin grup
// Parameters:
// - groupJID: JID grup
// Returns:
// - true jika pengguna adalah admin
// - error jika gagal
func (c *Client) IsGroupAdmin(groupJID types.JID) (bool, error) {
	if err := c.validateConnection(); err != nil {
		return false, err
	}

	// Dapatkan ID sendiri
	selfID := c.GetSelfID()
	if selfID == nil {
		return false, whatsapp.ErrSelfIDNotAvailable
	}

	// Dapatkan anggota grup
	participants, err := c.GetGroupParticipants(groupJID)
	if err != nil {
		return false, err
	}

	// Cari diri sendiri dalam daftar anggota
	participant, _, found := c.findParticipantByJID(participants, *selfID)
	if !found {
		return false, fmt.Errorf("pengguna tidak ditemukan dalam grup")
	}

	return participant.IsAdmin, nil
}

// GetGroupName mendapatkan nama grup dari ID grup
// Parameters:
// - groupID: ID grup
// Returns:
// - nama grup
// - error jika gagal
func (c *Client) GetGroupName(groupID string) (string, error) {
	group, err := c.GetGroupByID(groupID)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}

// GetGroupParticipantName mendapatkan nama anggota grup
// Parameters:
// - groupJID: JID grup
// - participantJID: JID anggota
// Returns:
// - nama anggota
// - error jika gagal
func (c *Client) GetGroupParticipantName(groupJID types.JID, participantJID types.JID) (string, error) {
	participants, err := c.GetGroupParticipants(groupJID)
	if err != nil {
		return "", err
	}

	participant, _, found := c.findParticipantByJID(participants, participantJID)
	if !found {
		return "", fmt.Errorf("anggota tidak ditemukan dalam grup")
	}

	if participant.DisplayName != "" {
		return participant.DisplayName, nil
	}

	// Gunakan langsung fungsi GetContactName
	return c.GetContactName(participant.JID, ""), nil
}
