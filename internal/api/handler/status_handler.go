package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// StatusHandler menangani endpoint status API
type StatusHandler struct {
	BaseHandler
	version string
}

// NewStatusHandler membuat instance baru StatusHandler
func NewStatusHandler(whatsClient *client.Client) *StatusHandler {
	return &StatusHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-status"),
		version:     "1.0.0",
	}
}

// GetStatus mengembalikan status koneksi WhatsApp
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	// Log untuk debugging
	h.Logger.Debug("GetStatus dipanggil", utils.Fields{
		"path": c.Path(),
	})

	// Periksa apakah whatsApp client nil
	if h.WhatsApp == nil {
		h.Logger.Error("whatsApp client is nil in GetStatus")
		return h.SendError(c, "WhatsApp client not initialized", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan status koneksi dengan pengecekan error
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		h.Logger.WithError(err).Error("Gagal mendapatkan status koneksi")
		return h.SendError(c, "Failed to get connection state", err, fiber.StatusInternalServerError)
	}

	// Konversi ke model dengan menyesuaikan data berdasarkan status koneksi
	status := model.ConnectionStatus{
		Status:      string(state.Status),
		IsConnected: state.IsConnected,
	}

	// Hanya sertakan informasi detail jika terhubung
	if state.IsConnected {
		// Gunakan format waktu Indonesia
		status.ConnectionRetries = state.ConnectionRetries
		status.LastActivity = utils.FormatTimeIndonesia(&state.LastActivity)
		status.Timestamp = utils.FormatTimeIndonesia(&state.Timestamp)
	}

	now := time.Now()
	response := model.NewStatusResponse("Status koneksi WhatsApp", status)
	response.Time = utils.FormatTimeIndonesia(&now)
	response.ServerTime = now

	return h.SendSuccess(c, response)
}

// TestConnection menguji koneksi API tanpa autentikasi
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	pingResponse := model.NewPingResponse("API berfungsi dengan baik", h.version)

	return h.SendSuccess(c, pingResponse)
}
