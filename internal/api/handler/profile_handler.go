package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
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

	// 3. Dapatkan status koneksi
	state := h.WhatsApp.GetConnectionState()

	// 4. Dapatkan informasi kontak dasar
	selfID := h.WhatsApp.GetSelfID()
	phoneNumber := ""
	if selfID != nil {
		phoneNumber = utils.FormatWhatsAppNumber(selfID.String())
	}

	// 5. Dapatkan nama kontak menggunakan service layer
	name := phoneNumber // Default ke nomor telepon
	if selfID != nil {
		// Gunakan fungsi GetContactName dari service
		contactName := h.WhatsApp.GetContactName(*selfID, phoneNumber)
		if contactName != "" {
			name = contactName
		}
	}

	// 6. Coba dapatkan URL foto profil jika tersedia
	pictureURL := ""
	profilePic, _ := h.WhatsApp.GetContactPictureURL(nil)
	if profilePic != "" {
		pictureURL = profilePic
	}

	// 7. Siapkan profil dengan informasi kontak saja (tanpa info perangkat)
	profile := model.ProfileInfo{
		IsConnected:    state.IsConnected,
		IsLoggedIn:     h.WhatsApp.IsLoggedIn(),
		ConnectedSince: h.FormatConnectedSince(state),
		PhoneNumber:    phoneNumber,
		Name:           name,
		PictureURL:     pictureURL,
	}

	// 8. Log dan kirim respons sukses
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
