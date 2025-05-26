package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// ProfileHandler menangani endpoint profil API
type ProfileHandler struct {
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewProfileHandler membuat instance baru ProfileHandler
func NewProfileHandler(whatsClient *client.Client) *ProfileHandler {
	return &ProfileHandler{
		whatsApp: whatsClient,
		logger:   utils.ForModule("handler-profile"),
	}
}

// GetProfile mengembalikan informasi profil akun WhatsApp terhubung
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	// Periksa koneksi
	if !h.whatsApp.GetConnectionState().IsConnected {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			model.NewProfileErrorResponse("WhatsApp tidak terhubung"))
	}

	// Dapatkan informasi perangkat/akun lengkap
	deviceInfo := h.whatsApp.GetDeviceInfo()

	// Format respons
	profile := model.ProfileInfo{
		IsConnected: h.whatsApp.GetConnectionState().IsConnected,
		IsLoggedIn:  h.whatsApp.IsLoggedIn(),
	}

	// Dapatkan informasi kontak sendiri untuk informasi yang lebih lengkap
	var contactInfo *types.ContactInfo
	if h.whatsApp.IsLoggedIn() {
		info, err := h.whatsApp.GetOwnContactInfo()
		if err == nil {
			contactInfo = info
			h.logger.Debug("Berhasil mendapatkan informasi kontak sendiri")
		} else {
			h.logger.Debug("Gagal mendapatkan informasi kontak sendiri", utils.Fields{
				"error": err.Error(),
			})
		}
	}

	// Isi data dari deviceInfo yang sudah lengkap
	if id, ok := deviceInfo["id"].(string); ok {
		profile.ID = id
		profile.PhoneNumber = client.FormatWhatsAppNumber(id)
	}

	// Ambil nama dari berbagai sumber dengan prioritas
	if name, ok := deviceInfo["push_name"].(string); ok && name != "" {
		profile.Name = name
	} else if contactInfo != nil && contactInfo.PushName != "" {
		profile.Name = contactInfo.PushName
	} else if name, ok := deviceInfo["name"].(string); ok && name != "" {
		profile.Name = name
	} else {
		// Fallback ke nomor telepon sebagai nama
		profile.Name = profile.PhoneNumber
	}

	// Ambil status dari deviceInfo
	if status, ok := deviceInfo["status"].(string); ok && status != "" {
		profile.Status = status
	}

	// Ambil informasi perangkat dari deviceInfo
	if device, ok := deviceInfo["device"].(string); ok && device != "" {
		profile.Device = device
	} else if platform, ok := deviceInfo["platform"].(string); ok && platform != "" {
		profile.Device = platform
	} else {
		profile.Device = "WhatsApp Web"
	}

	// Ambil URL foto profil dari deviceInfo
	if pictureURL, ok := deviceInfo["picture_url"].(string); ok && pictureURL != "" {
		profile.PictureURL = pictureURL
	}

	h.logger.Info("Informasi profil WhatsApp berhasil diambil", utils.Fields{
		"name":    profile.Name,
		"number":  profile.PhoneNumber,
		"device":  profile.Device,
		"has_pic": profile.PictureURL != "",
	})

	return c.JSON(model.NewProfileResponse("Profil WhatsApp berhasil diambil", profile))
}
