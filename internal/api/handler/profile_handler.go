package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ProfileHandler menangani endpoint informasi profil WhatsApp dengan dukungan multi-user
type ProfileHandler struct {
	BaseHandler
}

// NewProfileHandler membuat instance baru ProfileHandler
func NewProfileHandler(userManager *manager.UserManager) *ProfileHandler {
	return &ProfileHandler{
		BaseHandler: NewBaseHandler(userManager, "handler-profile"),
	}
}

// GetProfile mengembalikan informasi profil WhatsApp untuk pengguna tertentu
// Endpoint: GET /api/profile
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetProfile")

	// 2. Dapatkan client untuk user ini
	whatsClient, userID, err := h.GetUserClient(c)
	if err != nil {
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

	// 3. Validasi koneksi WhatsApp
	state, err := whatsClient.GetConnectionStateSafe()
	if err != nil || !state.IsConnected {
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

	// 4. Dapatkan informasi perangkat dan status koneksi
	deviceInfo := whatsClient.GetConnectionInfo()

	// 5. Persiapkan data profil
	profile := model.ProfileInfo{
		IsConnected:    state.IsConnected,
		IsLoggedIn:     whatsClient.IsLoggedIn(),
		ConnectedSince: h.FormatConnectedSince(state),
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
	if whatsClient.IsLoggedIn() && whatsClient.GetSelfID() != nil {
		if pictureURL, err := whatsClient.GetProfilePictureURL(); err == nil && pictureURL != "" {
			profile.PictureURL = pictureURL
		}
	}

	// Format nomor telepon untuk tampilan
	formattedNumber := utils.FormatWhatsAppNumber(profile.PhoneNumber)

	// 7. Log dan kirim respons sukses
	h.LogSuccessResponse("Informasi profil WhatsApp berhasil diambil", utils.Fields{
		"userID":      userID,
		"phone":       formattedNumber,
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