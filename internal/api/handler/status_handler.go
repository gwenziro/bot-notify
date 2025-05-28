package handler

import (
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

	// Buat respons dengan status yang sesuai
	message := "Status koneksi WhatsApp"
	if !state.IsConnected {
		message = "WhatsApp sedang tidak terhubung"
	}

	response := model.NewStatusResponse(message, string(state.Status), state.IsConnected)

	// Hanya sertakan informasi detail jika terhubung
	if state.IsConnected {
		response.ConnectionRetries = state.ConnectionRetries
		response.LastActivity = utils.FormatTimeIndonesia(&state.LastActivity)
	}

	// Sesuaikan kode status HTTP
	statusCode := fiber.StatusOK
	if !state.IsConnected {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

// TestConnection menguji koneksi API tanpa autentikasi
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	pingResponse := model.NewPingResponse("API berfungsi dengan baik", h.version)

	return h.SendSuccess(c, pingResponse)
}
