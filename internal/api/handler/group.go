package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// GroupHandler menangani endpoint grup API
type GroupHandler struct {
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewGroupHandler membuat instance baru GroupHandler
func NewGroupHandler(whatsClient *client.Client) *GroupHandler {
	return &GroupHandler{
		whatsApp: whatsClient,
		logger:   utils.ForModule("handler-group"),
	}
}

// ListGroups mengembalikan daftar grup yang tersedia
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	// Dapatkan JID perangkat kita sendiri untuk membandingkan dengan admin grup
	selfJID := h.whatsApp.GetSelfID()
	if selfJID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewGroupListResponse("Gagal mendapatkan ID perangkat", nil))
	}

	// Gunakan WhatsApp client langsung untuk mendapatkan daftar grup
	groups, err := h.whatsApp.GetGroups()
	if err != nil {
		h.logger.WithError(err).Error("Gagal mendapatkan daftar grup")
		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewGroupListResponse("Gagal mendapatkan daftar grup: "+err.Error(), nil))
	}

	// Tambahkan log untuk debugging
	h.logger.Debug("Self JID for admin check", utils.Fields{
		"jid": selfJID.String(),
	})

	// Konversi ke bentuk yang sesuai untuk respons API
	result := make([]model.GroupInfo, len(groups))
	for i, group := range groups {
		// Dapatkan informasi grup lebih detail
		detailedGroup, err := h.whatsApp.GetGroupByID(group.JID.String())
		if err != nil {
			h.logger.WithError(err).Warn("Gagal mendapatkan informasi detail grup", utils.Fields{
				"group_id": group.JID.String(),
			})
			// Gunakan informasi grup yang sudah ada jika detail gagal diambil
			detailedGroup = group
		}

		// Periksa apakah perangkat kita adalah admin grup menggunakan data participant
		isAdmin := false
		normalizedSelfJID := normalizeJID(selfJID.String())

		// Konversi participants ke model
		participants := make([]model.GroupParticipantInfo, len(detailedGroup.Participants))
		for j, participant := range detailedGroup.Participants {
			// Normalize participant JID untuk perbandingan
			normalizedParticipantJID := normalizeJID(participant.JID.String())

			// Periksa apakah ini adalah JID kita
			if normalizedSelfJID == normalizedParticipantJID {
				isAdmin = participant.IsAdmin
				h.logger.Debug("Found self in participants", utils.Fields{
					"is_admin": participant.IsAdmin,
					"jid":      normalizedParticipantJID,
				})
			}

			// Tambahkan ke daftar participants
			participants[j] = model.GroupParticipantInfo{
				JID:          participant.JID.String(),
				PhoneNumber:  client.FormatWhatsAppNumber(participant.JID.String()),
				IsAdmin:      participant.IsAdmin,
				IsSuperAdmin: participant.IsSuperAdmin,
				DisplayName:  participant.DisplayName,
			}
		}

		result[i] = model.GroupInfo{
			ID:           group.JID.String(),
			Name:         group.Name,
			MemberCount:  len(group.Participants),
			IsAdmin:      isAdmin,
			Participants: participants,
		}
	}

	h.logger.WithField("count", len(groups)).Info("Daftar grup berhasil diambil")

	// Kirim response sukses menggunakan model terkait
	return c.JSON(model.NewGroupListResponse("Daftar grup berhasil diambil", result))
}

// normalizeJID menormalkan JID untuk perbandingan yang lebih andal
func normalizeJID(jid string) string {
	// Hapus bagian server dan device identifier jika ada
	normalized := jid

	// Hapus bagian setelah @s.whatsapp.net atau @g.us
	if idx := strings.IndexByte(normalized, '@'); idx > 0 {
		// Ekstrak bagian nomor saja
		userPart := normalized[:idx]
		// Ambil bagian "server" (s.whatsapp.net atau g.us)
		serverPart := ""
		if strings.Contains(normalized, "@s.whatsapp.net") {
			serverPart = "@s.whatsapp.net"
		} else if strings.Contains(normalized, "@g.us") {
			serverPart = "@g.us"
		}

		// Hapus device identifier (misalnya ":3" pada "123456789:3@s.whatsapp.net")
		if deviceIdx := strings.IndexByte(userPart, ':'); deviceIdx > 0 {
			userPart = userPart[:deviceIdx]
		}

		// Gabungkan kembali
		normalized = userPart + serverPart
	}

	return normalized
}
