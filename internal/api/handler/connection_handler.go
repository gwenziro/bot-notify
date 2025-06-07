package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ConnectionHandler menangani endpoint manajemen koneksi WhatsApp
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
// Endpoint: POST /api/reconnect
func (h *ConnectionHandler) Reconnect(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Reconnect")

	// 2. Parse request jika ada
	var req model.ReconnectRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		// Abaikan error parsing karena parameter opsional
		h.Logger.Debug("Gagal parsing request body, melanjutkan dengan nilai default")
	}

	// 3. Hubungkan WhatsApp
	err := h.WhatsApp.Connect()
	if err != nil {
		return h.SendError(c, constants.MsgConnectionFailed, err, fiber.StatusInternalServerError)
	}

	// 4. Log dan kirim respons sukses
	h.LogSuccessResponse("Permintaan menghubungkan ulang WhatsApp berhasil diproses", utils.Fields{
		"force":  req.Force,
		"status": string(h.WhatsApp.GetConnectionState().Status),
	})

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgReconnectSuccess,
		string(h.WhatsApp.GetConnectionState().Status)))
}

// Disconnect memutuskan koneksi WhatsApp
// Endpoint: POST /api/disconnect
func (h *ConnectionHandler) Disconnect(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Disconnect")

	// 2. Putuskan koneksi WhatsApp
	h.WhatsApp.Disconnect()

	// 3. Tunggu sejenak agar status koneksi diperbarui
	time.Sleep(300 * time.Millisecond)

	// 4. Hapus sesi
	err := h.WhatsApp.SessionManager.ClearSessions()
	if err != nil {
		return h.SendError(c, constants.MsgFailedSessionDelete, err, fiber.StatusInternalServerError)
	}

	// 5. Verifikasi status koneksi
	state := h.WhatsApp.GetConnectionState()
	if state.IsConnected {
		h.Logger.Warn("Status koneksi masih terdeteksi sebagai terhubung setelah disconnect")
	}

	// 6. Log dan kirim respons sukses
	h.LogSuccessResponse("WhatsApp berhasil diputuskan melalui API", utils.Fields{
		"status": string(state.Status),
	})

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgDisconnectSuccess,
		string(state.Status)))
}
