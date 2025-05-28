package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// GroupHandler menangani endpoint grup API
type GroupHandler struct {
	BaseHandler
}

// NewGroupHandler membuat instance baru GroupHandler
func NewGroupHandler(whatsClient *client.Client) *GroupHandler {
	return &GroupHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-group"),
	}
}

// ListGroups mengembalikan daftar grup yang tersedia
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, "WhatsApp sedang tidak terhubung") {
		// Karena ListGroups mengembalikan array kosong saat tidak terhubung, kita perlu membuat respons khusus
		now := time.Now()
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.GroupListResponse{
			BaseResponse: model.BaseResponse{
				Success: false,
				Message: "WhatsApp sedang tidak terhubung",
				Time:    utils.FormatTimeIndonesia(&now),
			},
			Count:  0,
			Groups: []model.GroupInfo{},
		})
	}

	// Dapatkan JID perangkat sendiri
	selfJID := h.WhatsApp.GetSelfID()
	if selfJID == nil {
		return h.SendError(c, "Gagal mendapatkan ID perangkat", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan daftar grup
	groups, err := h.WhatsApp.GetGroups()
	if err != nil {
		h.Logger.WithError(err).Error("Gagal mendapatkan daftar grup")
		return h.SendError(c, "Gagal mendapatkan daftar grup", err, fiber.StatusInternalServerError)
	}

	// Proses data grup
	result := h.processGroups(groups, selfJID)

	h.Logger.WithField("count", len(groups)).Info("Daftar grup berhasil diambil")
	return h.SendSuccess(c, model.NewGroupListResponse("Daftar grup berhasil diambil", result))
}

// processGroups mengkonversi daftar grup WhatsApp ke model API
func (h *GroupHandler) processGroups(groups []*types.GroupInfo, selfJID *types.JID) []model.GroupInfo {
	result := make([]model.GroupInfo, len(groups))
	normalizedSelfJID := utils.NormalizeJID(selfJID.String())

	for i, group := range groups {
		// Dapatkan informasi detail grup
		detailedGroup, err := h.WhatsApp.GetGroupByID(group.JID.String())
		if err != nil {
			h.Logger.WithError(err).Warn("Gagal mendapatkan informasi detail grup", utils.Fields{
				"group_id": group.JID.String(),
			})
			detailedGroup = group
		}

		// Mendapatkan partisipan dengan info yang diperkaya
		enrichedParticipants, err := h.WhatsApp.GetEnrichedParticipants(group.JID)
		if err != nil {
			h.Logger.WithError(err).Warn("Gagal mendapatkan informasi kontak partisipan untuk daftar grup", utils.Fields{
				"group_id": group.JID.String(),
			})
			// Tetap gunakan participants normal jika gagal mendapatkan yang diperkaya
			enrichedParticipants = detailedGroup.Participants
		}

		// Proses partisipan dan cek apakah pengguna adalah admin
		participants, isAdmin := h.processParticipantsWithContacts(enrichedParticipants, normalizedSelfJID)

		result[i] = model.GroupInfo{
			ID:           group.JID.String(),
			Name:         group.Name,
			MemberCount:  len(detailedGroup.Participants),
			IsAdmin:      isAdmin,
			Participants: participants,
		}
	}

	return result
}

// GetParticipants mengembalikan daftar partisipan dari sebuah grup
func (h *GroupHandler) GetParticipants(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, "Gagal mendapatkan daftar anggota grup: WhatsApp sedang tidak terhubung") {
		return nil
	}

	// Dapatkan groupID dari parameter
	groupID := c.Params("id")
	if groupID == "" {
		return h.SendError(c, "ID grup harus disediakan", nil, fiber.StatusBadRequest)
	}

	// Validasi format ID grup
	if !utils.ValidateGroupID(groupID) {
		return h.SendError(c, "Format ID grup tidak valid", nil, fiber.StatusBadRequest)
	}

	// Dapatkan JID perangkat sendiri
	selfJID := h.WhatsApp.GetSelfID()
	if selfJID == nil {
		return h.SendError(c, "Gagal mendapatkan ID perangkat", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan informasi grup
	group, err := h.WhatsApp.GetGroupByID(groupID)
	if err != nil {
		h.Logger.WithError(err).Error("Gagal mendapatkan informasi grup", utils.Fields{
			"group_id": groupID,
		})
		return h.SendError(c, "Gagal mendapatkan informasi grup", err, fiber.StatusNotFound)
	}

	// Mendapatkan partisipan dengan info yang diperkaya
	enrichedParticipants, err := h.WhatsApp.GetEnrichedParticipants(group.JID)
	if err != nil {
		h.Logger.WithError(err).Warn("Gagal mendapatkan informasi kontak partisipan", utils.Fields{
			"group_id": groupID,
		})
		// Tetap gunakan participants normal jika gagal mendapatkan yang diperkaya
		enrichedParticipants = group.Participants
	}

	// Proses partisipan dan cek apakah pengguna adalah admin
	normalizedSelfJID := utils.NormalizeJID(selfJID.String())
	participants, isAdmin := h.processParticipantsWithContacts(enrichedParticipants, normalizedSelfJID)

	h.Logger.WithFields(utils.Fields{
		"group_id": groupID,
		"count":    len(participants),
		"is_admin": isAdmin,
	}).Info("Daftar anggota grup berhasil diambil")

	return h.SendSuccess(c, model.NewGroupParticipantsResponse(
		"Daftar anggota grup berhasil diambil",
		group.JID.String(),
		group.Name,
		isAdmin,
		participants,
	))
}

// processParticipantsWithContacts memproses partisipan grup dengan informasi kontak lengkap
func (h *GroupHandler) processParticipantsWithContacts(participants []types.GroupParticipant, normalizedSelfJID string) ([]model.GroupParticipantInfo, bool) {
	result := make([]model.GroupParticipantInfo, len(participants))
	isAdmin := false

	for i, participant := range participants {
		normalizedParticipantJID := utils.NormalizeJID(participant.JID.String())

		// Periksa apakah ini adalah JID kita
		if normalizedSelfJID == normalizedParticipantJID {
			isAdmin = participant.IsAdmin
		}

		// Format nomor telepon untuk display
		phoneNumber := client.FormatWhatsAppNumber(participant.JID.String())

		// Pilih nama kontak yang terbaik untuk ditampilkan
		contactName := participant.DisplayName
		pushName := "" // Simpan pushName terpisah jika tersedia dari store

		// Tentukan nama kontak terbaik untuk ditampilkan
		displayName := contactName
		if displayName == "" {
			displayName = phoneNumber // Fallback ke nomor jika tidak ada nama
		}

		// Tambahkan ke model
		result[i] = model.GroupParticipantInfo{
			JID:          participant.JID.String(),
			PhoneNumber:  phoneNumber,
			IsAdmin:      participant.IsAdmin,
			IsSuperAdmin: participant.IsSuperAdmin,
			DisplayName:  displayName,
			PushName:     pushName,
			ContactName:  contactName,
		}
	}

	return result, isAdmin
}
