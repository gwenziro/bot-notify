package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ConnectionHandler menangani endpoint manajemen koneksi WhatsApp dengan dukungan multi-user
type ConnectionHandler struct {
	BaseHandler
}

// NewConnectionHandler membuat instance baru ConnectionHandler
func NewConnectionHandler(userManager *manager.UserManager) *ConnectionHandler {
	return &ConnectionHandler{
		BaseHandler: NewBaseHandler(userManager, "handler-connection"),
	}
}

// Reconnect mencoba menghubungkan ulang WhatsApp untuk pengguna tertentu
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

	// 3. Dapatkan client untuk user ini
	whatsClient, userID, err := h.GetUserClient(c)
	if err != nil {
		return h.SendError(c, constants.MsgClientNotAvailable, err, fiber.StatusInternalServerError)
	}

	// 4. Hubungkan WhatsApp
	err = whatsClient.Connect()
	if err != nil {
		return h.SendError(c, constants.MsgConnectionFailed, err, fiber.StatusInternalServerError)
	}

	// 5. Log dan kirim respons sukses
	h.LogSuccessResponse("Permintaan menghubungkan ulang WhatsApp berhasil diproses", utils.Fields{
		"userID": userID,
		"force":  req.Force,
		"status": string(whatsClient.GetConnectionState().Status),
	})

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgReconnectSuccess,
		string(whatsClient.GetConnectionState().Status)))
}

// Disconnect memutuskan koneksi WhatsApp untuk pengguna tertentu
// Endpoint: POST /api/disconnect
func (h *ConnectionHandler) Disconnect(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Disconnect")

	// 2. Dapatkan client untuk user ini
	whatsClient, userID, err := h.GetUserClient(c)
	if err != nil {
		return h.SendError(c, constants.MsgClientNotAvailable, err, fiber.StatusInternalServerError)
	}

	// 3. Putuskan koneksi WhatsApp
	whatsClient.Disconnect()

	// 4. Tunggu sejenak agar status koneksi diperbarui
	time.Sleep(300 * time.Millisecond)

	// 5. Hapus sesi
	err = whatsClient.SessionManager.ClearSessions()
	if err != nil {
		return h.SendError(c, constants.MsgFailedSessionDelete, err, fiber.StatusInternalServerError)
	}

	// 6. Verifikasi status koneksi
	state := whatsClient.GetConnectionState()
	if state.IsConnected {
		h.Logger.Warn("Status koneksi masih terdeteksi sebagai terhubung setelah disconnect", utils.Fields{
			"userID": userID,
		})
	}

	// 7. Log dan kirim respons sukses
	h.LogSuccessResponse("WhatsApp berhasil diputuskan melalui API", utils.Fields{
		"userID": userID,
		"status": string(state.Status),
	})

	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgDisconnectSuccess,
		string(state.Status)))
}