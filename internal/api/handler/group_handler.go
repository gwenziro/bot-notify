package handler

import (
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
	// Dapatkan status koneksi terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas tanpa data sensitif
	if !state.IsConnected {
		h.Logger.Info("Permintaan daftar grup saat WhatsApp tidak terhubung")
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.NewGroupListResponse("WhatsApp sedang tidak terhubung", []model.GroupInfo{}))
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

		// Proses partisipan dan cek apakah pengguna adalah admin
		participants, isAdmin := h.processParticipants(detailedGroup.Participants, normalizedSelfJID)

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

// processParticipants memproses partisipan grup dan memeriksa status admin
func (h *GroupHandler) processParticipants(participants []types.GroupParticipant, normalizedSelfJID string) ([]model.GroupParticipantInfo, bool) {
	result := make([]model.GroupParticipantInfo, len(participants))
	isAdmin := false

	for i, participant := range participants {
		normalizedParticipantJID := utils.NormalizeJID(participant.JID.String())

		// Periksa apakah ini adalah JID kita
		if normalizedSelfJID == normalizedParticipantJID {
			isAdmin = participant.IsAdmin
		}

		// Tambahkan ke model
		result[i] = model.GroupParticipantInfo{
			JID:          participant.JID.String(),
			PhoneNumber:  client.FormatWhatsAppNumber(participant.JID.String()),
			IsAdmin:      participant.IsAdmin,
			IsSuperAdmin: participant.IsSuperAdmin,
			DisplayName:  participant.DisplayName,
		}
	}

	return result, isAdmin
}
