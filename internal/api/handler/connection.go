package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ConnectionHandler menangani endpoint koneksi WhatsApp
type ConnectionHandler struct {
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewConnectionHandler membuat instance baru ConnectionHandler
func NewConnectionHandler(whatsClient *client.Client) *ConnectionHandler {
	return &ConnectionHandler{
		whatsApp: whatsClient,
		logger:   utils.ForModule("handler-connection"),
	}
}

// Reconnect mencoba menghubungkan ulang WhatsApp
func (h *ConnectionHandler) Reconnect(c *fiber.Ctx) error {
	var req model.ReconnectRequest

	// Parse request jika ada
	c.BodyParser(&req)

	// Connect whatsapp langsung
	err := h.whatsApp.Connect()
	if err != nil {
		h.logger.WithError(err).Error("Gagal menghubungkan ulang WhatsApp")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"sukses": false,
			"pesan":  "Gagal menghubungkan WhatsApp: " + err.Error(),
			"waktu":  utils.FormatTime(nil),
		})
	}

	h.logger.Info("Permintaan menghubungkan ulang WhatsApp berhasil diproses")

	return c.JSON(model.NewConnectionResponse(
		true,
		"Permintaan menghubungkan ulang WhatsApp berhasil diproses",
		"connecting"))
}

// Disconnect memutuskan koneksi WhatsApp
func (h *ConnectionHandler) Disconnect(c *fiber.Ctx) error {
	// Putuskan koneksi WhatsApp
	h.whatsApp.Disconnect()

	// Hapus sesi
	err := h.whatsApp.SessionManager.ClearSessions()
	if err != nil {
		h.logger.WithError(err).Error("Gagal menghapus sesi WhatsApp")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"sukses": false,
			"pesan":  "Koneksi diputus tetapi gagal menghapus sesi: " + err.Error(),
			"waktu":  utils.FormatTime(nil),
		})
	}

	h.logger.Info("WhatsApp berhasil diputuskan melalui API")

	return c.JSON(model.NewConnectionResponse(
		true,
		"WhatsApp berhasil diputuskan dan sesi dibersihkan",
		"disconnected"))
}
