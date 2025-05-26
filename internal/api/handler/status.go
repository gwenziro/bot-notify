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
	whatsApp *client.Client
	logger   utils.LogrusEntry
	version  string
}

// NewStatusHandler membuat instance baru StatusHandler
func NewStatusHandler(whatsClient *client.Client) *StatusHandler {
	logger := utils.ForModule("handler-status")

	if whatsClient == nil {
		logger.Error("whatsClient tidak boleh nil saat membuat StatusHandler")
	}

	return &StatusHandler{
		whatsApp: whatsClient,
		logger:   logger,
		version:  "1.0.0", // Versi API
	}
}

// GetStatus mengembalikan status koneksi WhatsApp
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	// Log untuk debugging
	h.logger.Debug("GetStatus dipanggil", utils.Fields{
		"has_client": h.whatsApp != nil,
		"path":       c.Path(),
	})

	// Periksa apakah whatsApp client nil
	if h.whatsApp == nil {
		h.logger.Error("whatsApp client is nil in GetStatus")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "WhatsApp client not initialized",
			"code":    fiber.StatusInternalServerError,
		})
	}

	// Dapatkan status koneksi dengan pengecekan error
	state, err := h.whatsApp.GetConnectionStateSafe()
	if err != nil {
		h.logger.WithError(err).Error("Gagal mendapatkan status koneksi")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get connection state: " + err.Error(),
			"code":    fiber.StatusInternalServerError,
		})
	}

	// Konversi ke model
	status := model.ConnectionStatus{
		Status:            string(state.Status),
		IsConnected:       state.IsConnected,
		ConnectionRetries: state.ConnectionRetries,
		LastActivity:      state.LastActivity,
		Timestamp:         state.Timestamp,

		// Format waktu untuk tampilan yang lebih baik
		LastActivityFormatted: utils.FormatTimeShort(&state.LastActivity),
		TimestampFormatted:    utils.FormatTimeShort(&state.Timestamp),
	}

	now := time.Now()
	response := model.StatusResponse{
		Success: true,
		Status:  string(state.Status),
		Details: status,
		Time:    now,

		// Format waktu untuk tampilan yang lebih baik
		TimeFormatted: utils.FormatTimeShort(&now),
	}

	return c.JSON(response)
}

// TestConnection menguji koneksi API tanpa autentikasi
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	now := time.Now()
	pingResponse := model.PingResponse{
		Success: true,
		Message: "API berfungsi dengan baik",
		Time:    now,
		Version: h.version,

		// Format waktu untuk tampilan yang lebih baik
		TimeFormatted: utils.FormatTimeShort(&now),
	}

	return c.JSON(pingResponse)
}
