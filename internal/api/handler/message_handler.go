package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// MessageHandler menangani endpoint pesan API
type MessageHandler struct {
	BaseHandler
}

// NewMessageHandler membuat instance baru MessageHandler
func NewMessageHandler(whatsClient *client.Client) *MessageHandler {
	return &MessageHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-message"),
	}
}

// SendGroup mengirim pesan ke grup
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// 1. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	// 2. Parse dan validasi input
	var req model.GroupMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	// Validasi required fields
	if req.GroupID == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "groupID"), nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "message"), nil, fiber.StatusBadRequest)
	}

	// Validasi tambahan
	if !utils.ValidateGroupID(req.GroupID) {
		return h.SendError(c, constants.MsgInvalidGroupID, nil, fiber.StatusBadRequest)
	}

	// 3. Proses bisnis: kirim pesan
	jid := client.ParseGroupID(req.GroupID)
	sendTime, err := h.WhatsApp.SendMessage(jid, req.Message)
	if err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan grup")
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	// 4. Log sukses dan kirim respons
	h.Logger.Info("Pesan grup berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendGroupSuccess,
		jid.String(),
		"group",
		sendTime))
}

// SendPersonal mengirim pesan ke nomor personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// 1. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	// 2. Parse dan validasi input
	var req model.PersonalMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	// Validasi required fields
	if req.PhoneNumber == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "phoneNumber"), nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "message"), nil, fiber.StatusBadRequest)
	}

	// Validasi tambahan
	if !utils.ValidatePhoneNumber(req.PhoneNumber) {
		return h.SendError(c, constants.MsgInvalidPhoneNumber, nil, fiber.StatusBadRequest)
	}

	// Tambahkan context timeout
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()
	c.SetUserContext(ctx)

	// 3. Proses bisnis: kirim pesan
	jid := client.ParsePhoneNumber(req.PhoneNumber)
	sendTime, err := h.WhatsApp.SendMessage(jid, req.Message)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.Logger.Error("Timeout saat mengirim pesan personal")
			return h.SendError(c, constants.MsgTimeoutError, err, fiber.StatusGatewayTimeout)
		}
		h.Logger.WithError(err).Error("Gagal mengirim pesan personal")
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	// 4. Log sukses dan kirim respons
	h.Logger.Info("Pesan personal berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendSuccess,
		jid.String(),
		"personal",
		sendTime))
}
