package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// StatusHandler menangani endpoint status koneksi API
type StatusHandler struct {
	BaseHandler
	version string // Versi aplikasi untuk ping response
}

// NewStatusHandler membuat instance baru StatusHandler
func NewStatusHandler(whatsClient *client.Client) *StatusHandler {
	return &StatusHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-status"),
		version:     constants.DefaultAppVersion,
	}
}

// GetStatus mengembalikan status koneksi WhatsApp
// Endpoint: GET /api/status
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetStatus")

	// 2. Periksa ketersediaan WhatsApp client
	if h.WhatsApp == nil {
		return h.SendError(c, constants.MsgClientNotAvailable, nil, fiber.StatusInternalServerError) // Perbarui referensi
	}

	// 3. Dapatkan status koneksi
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, constants.MsgStatusFailed, err, fiber.StatusInternalServerError) // Perbarui referensi
	}

	// 4. Tentukan pesan yang sesuai
	var message string
	if !state.IsConnected {
		message = constants.MsgNotConnected // Perbarui referensi
	} else {
		message = constants.MsgConnected // Perbarui referensi
	}

	// 5. Buat respons
	response := model.NewStatusResponse(message, string(state.Status), state.IsConnected)

	// 6. Sertakan informasi tambahan jika terhubung
	if state.IsConnected {
		response.ConnectionRetries = state.ConnectionRetries
	}

	// 7. Selalu kembalikan StatusOK (200) agar dashboard tetap bisa menampilkan status
	statusCode := fiber.StatusOK

	// 8. Log hasil
	h.LogSuccessResponse("Status koneksi berhasil diambil", utils.Fields{
		"status":      state.Status,
		"connected":   state.IsConnected,
		"http_status": statusCode,
	})

	return c.Status(statusCode).JSON(response)
}

// TestConnection menguji koneksi API tanpa autentikasi
// Endpoint: GET /ping
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "TestConnection")

	// 2. Buat respons ping
	pingResponse := model.NewPingResponse("API berfungsi dengan baik", h.version)

	// 3. Log hasil
	h.LogSuccessResponse("Test koneksi berhasil", utils.Fields{
		"version": h.version,
	})

	// 4. Kirim respons
	return h.SendSuccess(c, pingResponse)
}
