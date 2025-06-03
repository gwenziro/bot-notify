package handler

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// MessageHandler menangani endpoint pengiriman pesan API
type MessageHandler struct {
	BaseHandler
}

// NewMessageHandler membuat instance baru MessageHandler
func NewMessageHandler(whatsClient *client.Client) *MessageHandler {
	return &MessageHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-message"),
	}
}

// SendGroup mengirim pesan ke grup WhatsApp
// Endpoint: POST /api/send/group
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "SendGroup")

	// 2. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	// 3. Parse dan validasi input
	var req model.GroupMessageRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Validasi field wajib
	if err := h.ValidateRequiredField(c, "groupID", req.GroupID); err != nil {
		return err
	}

	if err := h.ValidateRequiredField(c, "message", req.Message); err != nil {
		return err
	}

	// 5. Validasi format data
	if !utils.ValidateGroupID(req.GroupID) {
		return h.SendError(c, constants.MsgInvalidGroupID, nil, fiber.StatusBadRequest)
	}

	// 6. Proses pengiriman pesan
	groupID := utils.FormatGroupID(req.GroupID)
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return h.SendError(c, "Format ID grup tidak valid: "+err.Error(), err, fiber.StatusBadRequest)
	}

	sendTime, err := h.WhatsApp.SendMessage(jid, req.Message)
	if err != nil {
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	// 7. Log dan kirim respons sukses
	h.LogSuccessResponse("Pesan grup berhasil dikirim", utils.Fields{
		"group_id": req.GroupID,
		"msg_len":  len(req.Message),
	})

	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendGroupSuccess,
		jid.String(),
		"group",
		sendTime))
}

// SendPersonal mengirim pesan ke kontak personal WhatsApp
// Endpoint: POST /api/send/personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "SendPersonal")

	// 2. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	// 3. Parse dan validasi input
	var req model.PersonalMessageRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Validasi field wajib
	if err := h.ValidateRequiredField(c, "phoneNumber", req.PhoneNumber); err != nil {
		return err
	}

	if err := h.ValidateRequiredField(c, "message", req.Message); err != nil {
		return err
	}

	// 5. Validasi format data
	if !utils.ValidatePhoneNumber(req.PhoneNumber) {
		return h.SendError(c, constants.MsgInvalidPhoneNumber, nil, fiber.StatusBadRequest)
	}

	// 6. Setup timeout context
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()
	c.SetUserContext(ctx)

	// 7. Proses pengiriman pesan
	phoneNumber := utils.FormatPhoneNumber(req.PhoneNumber)
	jid, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
	if err != nil {
		return h.SendError(c, "Format nomor telepon tidak valid: "+err.Error(), err, fiber.StatusBadRequest)
	}

	sendTime, err := h.WhatsApp.SendMessage(jid, req.Message)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return h.SendError(c, constants.MsgTimeoutError, err, fiber.StatusGatewayTimeout)
		}
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	// 8. Log dan kirim respons sukses
	h.LogSuccessResponse("Pesan personal berhasil dikirim", utils.Fields{
		"phone":   req.PhoneNumber,
		"msg_len": len(req.Message),
	})

	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendSuccess,
		jid.String(),
		"personal",
		sendTime))
}
