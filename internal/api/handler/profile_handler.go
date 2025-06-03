package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ProfileHandler menangani endpoint informasi profil WhatsApp
type ProfileHandler struct {
	BaseHandler
}

// NewProfileHandler membuat instance baru ProfileHandler
func NewProfileHandler(whatsClient *client.Client) *ProfileHandler {
	return &ProfileHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-profile"),
	}
}

// GetProfile mengembalikan informasi profil akun WhatsApp terhubung
// Endpoint: GET /api/profile
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetProfile")

	// 2. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, "Gagal mendapatkan profil: "+constants.MsgNotConnected) {
		// Buat profile kosong dengan informasi minimal
		emptyProfile := model.ProfileInfo{
			IsConnected: false,
			IsLoggedIn:  false,
		}

		// Kembalikan respons dengan format yang konsisten
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.NewProfileResponse(
			false,                     // Set success ke false
			constants.MsgNotConnected, // Pesan yang lebih sesuai
			emptyProfile,
		))
	}

	// 3. Dapatkan informasi perangkat dan status koneksi
	deviceInfo := h.WhatsApp.GetDeviceInfo()
	state := h.WhatsApp.GetConnectionState()

	// 4. Persiapkan data profil
	profile := model.ProfileInfo{
		IsConnected: state.IsConnected,
		IsLoggedIn:  h.WhatsApp.IsLoggedIn(),
	}

	// 5. Isi data profil dari deviceInfo
	if id, ok := deviceInfo["id"].(string); ok {
		profile.ID = id
		profile.PhoneNumber = utils.FormatWhatsAppNumber(id)
	}

	if pushName, ok := deviceInfo["push_name"].(string); ok && pushName != "" {
		profile.Name = pushName
	} else {
		profile.Name = profile.PhoneNumber
	}

	if status, ok := deviceInfo["status"].(string); ok {
		profile.Status = status
	}

	// 6. Tambahkan URL foto profil jika tersedia
	if h.WhatsApp.IsLoggedIn() && h.WhatsApp.GetSelfID() != nil {
		if pictureURL, err := h.WhatsApp.GetProfilePictureURL(); err == nil && pictureURL != "" {
			profile.PictureURL = pictureURL
		}
	}

	// 7. Log dan kirim respons sukses
	h.LogSuccessResponse("Informasi profil WhatsApp berhasil diambil", utils.Fields{
		"phone":       profile.PhoneNumber,
		"name":        profile.Name,
		"connected":   profile.IsConnected,
		"has_picture": profile.PictureURL != "",
	})

	return h.SendSuccess(c, model.NewProfileResponse(
		true, // Tetap true untuk respons sukses
		constants.MsgProfileRetrieved,
		profile,
	))
}
