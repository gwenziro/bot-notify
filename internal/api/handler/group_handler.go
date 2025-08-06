package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// GroupHandler menangani endpoint grup WhatsApp API dengan dukungan multi-user
type GroupHandler struct {
	BaseHandler
}

// NewGroupHandler membuat instance baru GroupHandler
func NewGroupHandler(userManager *manager.UserManager) *GroupHandler {
	return &GroupHandler{
		BaseHandler: NewBaseHandler(userManager, "handler-group"),
	}
}

// ListGroups mengembalikan daftar grup WhatsApp yang tersedia untuk pengguna tertentu
// Endpoint: GET /api/groups
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "ListGroups")

	// Gunakan metode standar untuk memeriksa koneksi dan mendapatkan client
	whatsClient, connected := h.CheckWhatsAppConnection(c, constants.MsgNotConnected)
	if !connected {
		// Karena ListGroups mengembalikan array kosong saat tidak terhubung, kita perlu membuat respons khusus
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.NewGroupListResponse(
			false, // Set success ke false saat tidak terhubung
			constants.MsgNotConnected,
			[]model.GroupInfo{},
		))
	}

	// Dapatkan JID perangkat sendiri
	selfJID := whatsClient.GetSelfID()
	if selfJID == nil {
		return h.SendError(c, fmt.Sprintf(constants.MsgFindIDFailed, "self"), nil, fiber.StatusInternalServerError)
	}

	// Dapatkan daftar grup
	groups, err := whatsClient.GetGroups()
	if err != nil {
		return h.SendError(c, fmt.Sprintf(constants.MsgGroupDataRetrievalFailed, err), nil, fiber.StatusInternalServerError)
	}

	// Proses data grup
	result := h.processGroups(groups, selfJID, whatsClient)

	h.LogSuccessResponse("Daftar grup berhasil diambil", utils.Fields{
		"count": len(groups),
	})

	return h.SendSuccess(c, model.NewGroupListResponse(
		true, // Set success ke true saat berhasil
		constants.MsgGroupsRetrieved,
		result))
}

// GetParticipants mengembalikan daftar anggota grup WhatsApp untuk pengguna tertentu
// Endpoint: GET /api/groups/:id/participants
func (h *GroupHandler) GetParticipants(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetParticipants")

	// Gunakan metode standar untuk memeriksa koneksi dan mendapatkan client
	whatsClient, connected := h.CheckWhatsAppConnection(c, "Gagal mendapatkan daftar anggota grup: "+constants.MsgNotConnected)
	if !connected {
		return nil
	}

	// Dapatkan groupID dari parameter
	groupID := c.Params("id")
	if groupID == "" {
		return h.SendError(c, constants.MsgGroupIDRequired, nil, fiber.StatusBadRequest)
	}

	// Validasi format ID grup
	if !utils.IsValidGroupID(groupID) {
		return h.SendError(c, constants.MsgInvalidGroupID, nil, fiber.StatusBadRequest)
	}

	// Dapatkan JID perangkat sendiri
	selfJID := whatsClient.GetSelfID()
	if selfJID == nil {
		return h.SendError(c, fmt.Sprintf(constants.MsgFindIDFailed, "self"), nil, fiber.StatusInternalServerError)
	}

	// Dapatkan informasi grup
	group, err := whatsClient.GetGroupByID(groupID)
	if err != nil {
		return h.SendError(c, fmt.Sprintf(constants.MsgGroupDataRetrievalFailed, err), nil, fiber.StatusNotFound)
	}

	// Mendapatkan partisipan dengan info yang diperkaya
	enrichedParticipants, err := whatsClient.GetEnrichedParticipants(group.JID)
	if err != nil {
		h.Logger.WithError(err).Warn("Gagal mendapatkan informasi kontak partisipan", utils.Fields{
			"group_id": groupID,
		})
		// Tetap gunakan participants normal jika gagal mendapatkan yang diperkaya
		enrichedParticipants = group.Participants
	}

	// Proses partisipan dan cek apakah pengguna adalah admin
	normalizedSelfJID := utils.NormalizeJID(selfJID.String())
	participants, isAdmin := h.processParticipantsWithContacts(enrichedParticipants, normalizedSelfJID, whatsClient)

	h.LogSuccessResponse("Daftar anggota grup berhasil diambil", utils.Fields{
		"group_id": groupID,
		"count":    len(participants),
		"is_admin": isAdmin,
	})

	return h.SendSuccess(c, model.NewGroupParticipantsResponse(
		constants.MsgGroupParticipantsRetrieved,
		group.JID.String(),
		group.Name,
		isAdmin,
		participants,
	))
}

// processGroups mengkonversi daftar grup WhatsApp ke model API
func (h *GroupHandler) processGroups(groups []*types.GroupInfo, selfJID *types.JID, whatsClient interface{}) []model.GroupInfo {
	result := make([]model.GroupInfo, len(groups))
	normalizedSelfJID := utils.NormalizeJID(selfJID.String())

	// Cast whatsClient ke interface yang diperlukan
	client, ok := whatsClient.(interface {
		GetGroupByID(string) (*types.GroupInfo, error)
		GetEnrichedParticipants(types.JID) ([]types.GroupParticipant, error)
	})
	if !ok {
		h.Logger.Error("Client tidak mendukung operasi grup yang diperlukan")
		return result
	}

	for i, group := range groups {
		// Dapatkan informasi detail grup
		detailedGroup, err := client.GetGroupByID(group.JID.String())
		if err != nil {
			h.Logger.WithError(err).Warn("Gagal mendapatkan informasi detail grup", utils.Fields{
				"group_id": group.JID.String(),
			})
			detailedGroup = group
		}

		// Mendapatkan partisipan dengan info yang diperkaya
		enrichedParticipants, err := client.GetEnrichedParticipants(group.JID)
		if err != nil {
			h.Logger.WithError(err).Warn("Gagal mendapatkan informasi kontak partisipan untuk daftar grup", utils.Fields{
				"group_id": group.JID.String(),
			})
			// Tetap gunakan participants normal jika gagal mendapatkan yang diperkaya
			enrichedParticipants = detailedGroup.Participants
		}

		// Proses partisipan dan cek apakah pengguna adalah admin
		participants, isAdmin := h.processParticipantsWithContacts(enrichedParticipants, normalizedSelfJID, whatsClient)

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

// processParticipantsWithContacts memproses partisipan grup dengan informasi kontak
func (h *GroupHandler) processParticipantsWithContacts(participants []types.GroupParticipant, normalizedSelfJID string, whatsClient interface{}) ([]model.GroupParticipantInfo, bool) {
	result := make([]model.GroupParticipantInfo, len(participants))
	isAdmin := false

	for i, participant := range participants {
		normalizedParticipantJID := utils.NormalizeJID(participant.JID.String())

		// Periksa apakah ini adalah JID kita
		if normalizedSelfJID == normalizedParticipantJID {
			isAdmin = participant.IsAdmin
		}

		// Format nomor telepon untuk display
		phoneNumber := utils.FormatWhatsAppNumber(participant.JID.String())

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