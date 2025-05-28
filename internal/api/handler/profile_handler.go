package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
)

// ProfileHandler menangani endpoint profil API
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
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, "Gagal mendapatkan profil: "+constants.MsgNotConnected) {
		// Buat profile kosong dengan informasi minimal
		emptyProfile := model.ProfileInfo{
			IsConnected: false,
			IsLoggedIn:  false,
		}

		// Kembalikan respons dengan format yang konsisten dan success=false
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.NewProfileResponse(
			false,                     // Set success ke false
			constants.MsgNotConnected, // Pesan yang lebih sesuai
			emptyProfile,
		))
	}

	// Dapatkan informasi perangkat/akun
	deviceInfo := h.WhatsApp.GetDeviceInfo()

	// Dapatkan status koneksi
	state := h.WhatsApp.GetConnectionState()

	// Format respons - hanya sertakan ConnectedSince jika terhubung
	profile := model.ProfileInfo{
		IsConnected:    state.IsConnected,
		IsLoggedIn:     h.WhatsApp.IsLoggedIn(),
		ConnectedSince: h.FormatConnectedSince(state),
	}

	// Isi data dari deviceInfo
	if id, ok := deviceInfo["id"].(string); ok {
		profile.ID = id
		profile.PhoneNumber = client.FormatWhatsAppNumber(id)
	}

	// Dapatkan nama dari deviceInfo (paling terpercaya)
	if pushName, ok := deviceInfo["push_name"].(string); ok && pushName != "" {
		profile.Name = pushName
	} else {
		// Fallback ke nomor telepon jika tidak ada nama
		profile.Name = profile.PhoneNumber
	}

	// Ambil status jika tersedia
	if status, ok := deviceInfo["status"].(string); ok {
		profile.Status = status
	}

	// Ambil URL foto profil jika tersedia
	if h.WhatsApp.IsLoggedIn() && h.WhatsApp.GetSelfID() != nil {
		if pictureURL, err := h.WhatsApp.GetProfilePictureURL(); err == nil && pictureURL != "" {
			profile.PictureURL = pictureURL
		}
	}

	h.Logger.Info("Informasi profil WhatsApp berhasil diambil")
	return h.SendSuccess(c, model.NewProfileResponse(
		true, // Tetap true untuk respons sukses
		constants.MsgProfileRetrieved,
		profile,
	))
}
