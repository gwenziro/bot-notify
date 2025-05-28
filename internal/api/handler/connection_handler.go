package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
)

// ConnectionHandler menangani endpoint koneksi WhatsApp
type ConnectionHandler struct {
	BaseHandler
}

// NewConnectionHandler membuat instance baru ConnectionHandler
func NewConnectionHandler(whatsClient *client.Client) *ConnectionHandler {
	return &ConnectionHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-connection"),
	}
}

// Reconnect mencoba menghubungkan ulang WhatsApp
func (h *ConnectionHandler) Reconnect(c *fiber.Ctx) error {
	var req model.ReconnectRequest

	// Parse request jika ada
	c.BodyParser(&req)

	// Connect whatsapp langsung
	err := h.WhatsApp.Connect()
	if err != nil {
		h.Logger.WithError(err).Error("Gagal menghubungkan ulang WhatsApp")
		return h.SendError(c, constants.MsgConnectionFailed, err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Permintaan menghubungkan ulang WhatsApp berhasil diproses")

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgReconnectSuccess,
		string(h.WhatsApp.GetConnectionState().Status)))
}

// Disconnect memutuskan koneksi WhatsApp
func (h *ConnectionHandler) Disconnect(c *fiber.Ctx) error {
	// Putuskan koneksi WhatsApp terlebih dahulu
	h.WhatsApp.Disconnect()

	// Tunggu sejenak agar status koneksi sempat diperbarui
	time.Sleep(300 * time.Millisecond)

	// Hapus sesi
	err := h.WhatsApp.SessionManager.ClearSessions()
	if err != nil {
		h.Logger.WithError(err).Error("Gagal menghapus sesi WhatsApp")
		return h.SendError(c, constants.MsgFailedSessionDelete, err, fiber.StatusInternalServerError)
	}

	// Verifikasi status koneksi setelah disconnect
	state := h.WhatsApp.GetConnectionState()
	if state.IsConnected {
		h.Logger.Warn("Status koneksi masih terdeteksi sebagai terhubung setelah disconnect")
	}

	h.Logger.Info("WhatsApp berhasil diputuskan melalui API")

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgDisconnectSuccess,
		string(h.WhatsApp.GetConnectionState().Status)))
}
