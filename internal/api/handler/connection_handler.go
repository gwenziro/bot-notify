package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// ConnectionHandler menangani endpoint koneksi WhatsApp API
type ConnectionHandler struct {
	BaseHandler
}

// NewConnectionHandler membuat instance baru ConnectionHandler
func NewConnectionHandler(whatsClient *client.Client) *ConnectionHandler {
	return &ConnectionHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-connection"),
	}
}

// Reconnect mencoba menghubungkan ulang ke WhatsApp
// Endpoint: POST /api/reconnect
func (h *ConnectionHandler) Reconnect(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Reconnect")

	// 2. Validasi client tidak nil
	if h.WhatsApp == nil {
		return h.SendError(c, constants.MsgClientNotAvailable, nil, fiber.StatusInternalServerError)
	}

	// 3. Parse dan validasi input
	var req model.ReconnectRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Periksa status koneksi saat ini
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, constants.MsgStatusFailed, err, fiber.StatusInternalServerError)
	}

	// 5. Jika sudah terhubung dan tidak force, kembalikan status saat ini
	if state.IsConnected && !req.Force {
		h.Logger.Info("Mencoba reconnect saat sudah terhubung tanpa force")
		return h.SendSuccess(c, model.NewConnectionResponse(
			true,
			constants.MsgConnected,
			string(state.Status),
		))
	}

	// 6. Putuskan koneksi saat ini jika ada
	if state.IsConnected {
		h.Logger.Info("Memutuskan koneksi sebelum reconnect dengan force")
		h.WhatsApp.Disconnect()
	}

	// 7. Coba hubungkan ulang
	err = h.WhatsApp.Connect()
	if err != nil {
		h.Logger.WithError(err).Error("Gagal melakukan reconnect")
		return h.SendError(c, constants.MsgConnectionFailed, err, fiber.StatusInternalServerError)
	}

	h.LogSuccessResponse("Reconnect berhasil diproses", utils.Fields{
		"force": req.Force,
	})

	// 8. Kirim respons sukses
	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgReconnectSuccess,
		string(client.StatusConnecting),
	))
}

// Disconnect memutuskan koneksi WhatsApp
// Endpoint: POST /api/disconnect
func (h *ConnectionHandler) Disconnect(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Disconnect")

	// 2. Validasi client tidak nil
	if h.WhatsApp == nil {
		return h.SendError(c, constants.MsgClientNotAvailable, nil, fiber.StatusInternalServerError)
	}

	// 3. Periksa status koneksi saat ini
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, constants.MsgStatusFailed, err, fiber.StatusInternalServerError)
	}

	// 4. Jika sudah tidak terhubung, kembalikan status saat ini
	if !state.IsConnected {
		h.Logger.Info("Mencoba disconnect saat sudah tidak terhubung")
		return h.SendSuccess(c, model.NewConnectionResponse(
			true,
			constants.MsgNotConnected,
			string(state.Status),
		))
	}

	// 5. Putuskan koneksi
	h.WhatsApp.Disconnect()

	// 6. Coba hapus sesi jika diperlukan
	err = h.WhatsApp.SessionManager.ClearSessions()
	if err != nil {
		h.Logger.WithError(err).Warn("Gagal menghapus sesi WhatsApp saat disconnect")
		return h.SendSuccess(c, model.NewConnectionResponse(
			true,
			constants.MsgFailedSessionDelete,
			string(client.StatusDisconnected),
		))
	}

	h.LogSuccessResponse("Disconnect berhasil", utils.Fields{
		"cleared_sessions": true,
	})

	// 7. Kirim respons sukses
	return h.SendSuccess(c, model.NewConnectionResponse(
		true,
		constants.MsgDisconnectSuccess,
		string(client.StatusDisconnected),
	))
}
