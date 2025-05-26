package handler

import (
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
	// Gunakan WhatsApp client langsung untuk mendapatkan daftar grup
	groups, err := h.whatsApp.GetGroups()
	if err != nil {
		h.logger.WithError(err).Error("Gagal mendapatkan daftar grup")
		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewGroupListResponse("Gagal mendapatkan daftar grup: "+err.Error(), nil))
	}

	// Dapatkan JID perangkat kita sendiri untuk membandingkan dengan admin grup
	selfJID := h.whatsApp.GetSelfID()

	// Konversi ke bentuk yang sesuai untuk respons API
	result := make([]model.GroupInfo, len(groups))
	for i, group := range groups {
		// Periksa apakah perangkat kita adalah admin grup
		isAdmin := false
		if selfJID != nil {
			for _, participant := range group.Participants {
				if participant.JID.String() == selfJID.String() && participant.IsAdmin {
					isAdmin = true
					break
				}
			}
		}

		result[i] = model.GroupInfo{
			ID:          group.JID.String(),
			Name:        group.Name,
			MemberCount: len(group.Participants),
			IsAdmin:     isAdmin,
		}
	}

	h.logger.WithField("count", len(groups)).Info("Daftar grup berhasil diambil")

	// Kirim response sukses menggunakan model terkait
	return c.JSON(model.NewGroupListResponse("Daftar grup berhasil diambil", result))
}
