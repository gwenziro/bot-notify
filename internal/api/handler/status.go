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
	return &StatusHandler{
		whatsApp: whatsClient,
		logger:   utils.ForModule("handler-status"),
		version:  "1.0.0", // Versi API
	}
}

// GetStatus mengembalikan status koneksi WhatsApp
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	state := h.whatsApp.GetConnectionState()

	// Konversi ke model
	status := model.ConnectionStatus{
		Status:            string(state.Status),
		IsConnected:       state.IsConnected,
		ConnectionRetries: state.ConnectionRetries,
		LastActivity:      state.LastActivity,
		Timestamp:         state.Timestamp,
	}

	response := model.StatusResponse{
		Success: true,
		Status:  string(state.Status),
		Details: status,
		Time:    time.Now(),
	}

	return c.JSON(response)
}

// TestConnection menguji koneksi API tanpa autentikasi
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	pingResponse := model.PingResponse{
		Success: true,
		Message: "API berfungsi dengan baik",
		Time:    time.Now(),
		Version: h.version,
	}

	return c.JSON(pingResponse)
}
