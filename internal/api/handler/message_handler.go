package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// MessageHandler menangani endpoint API pesan WhatsApp dengan dukungan multi-user
type MessageHandler struct {
	BaseHandler
}

// NewMessageHandler membuat instance baru MessageHandler
func NewMessageHandler(userManager *manager.UserManager) *MessageHandler {
	return &MessageHandler{
		BaseHandler: NewBaseHandler(userManager, "handler-message"),
	}
}

// SendGroup mengirim pesan ke grup WhatsApp
// Endpoint: POST /api/send/group
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "SendGroup")

	// 2. Gunakan metode standar untuk memeriksa koneksi dan mendapatkan client
	whatsClient, connected := h.CheckWhatsAppConnection(c, constants.MsgNotConnected)
	if !connected {
		return nil
	}

	// 3. Parse dan validasi request
	var req model.GroupMessageRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Validasi parameter
	if err := h.ValidateRequiredField(c, "groupID", req.GroupID); err != nil {
		return err
	}
	if err := h.ValidateRequiredField(c, "message", req.Message); err != nil {
		return err
	}

	// 5. Validasi format ID grup
	if !utils.IsValidGroupID(req.GroupID) {
		return h.SendError(c, constants.MsgInvalidGroupID, nil, fiber.StatusBadRequest)
	}

	// 6. Parse ID grup ke JID
	jid := utils.ParseGroupID(req.GroupID)

	// 7. Kirim pesan
	h.Logger.Info("Mengirim pesan grup", utils.Fields{
		"group_id": req.GroupID,
		"length":   len(req.Message),
	})

	sentTime, err := whatsClient.SendMessage(jid, req.Message)
	if err != nil {
		return h.SendError(c, fmt.Sprintf("%s: %v", constants.MsgSendFailure, err), err, fiber.StatusInternalServerError)
	}

	// 8. Log sukses
	h.LogSuccessResponse("Pesan grup berhasil dikirim", utils.Fields{
		"group_id":  req.GroupID,
		"timestamp": sentTime.Format(time.RFC3339),
	})

	// 9. Kirim respons sukses
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendSuccess,
		req.GroupID,
		"group",
		sentTime,
	))
}

// SendPersonal mengirim pesan ke nomor WhatsApp personal
// Endpoint: POST /api/send/personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "SendPersonal")

	// 2. Gunakan metode standar untuk memeriksa koneksi dan mendapatkan client
	whatsClient, connected := h.CheckWhatsAppConnection(c, constants.MsgNotConnected)
	if !connected {
		return nil
	}

	// 3. Parse dan validasi request
	var req model.PersonalMessageRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Validasi parameter
	if err := h.ValidateRequiredField(c, "phoneNumber", req.PhoneNumber); err != nil {
		return err
	}
	if err := h.ValidateRequiredField(c, "message", req.Message); err != nil {
		return err
	}

	// 5. Validasi format nomor telepon
	if !utils.IsValidPhoneNumber(req.PhoneNumber) {
		return h.SendError(c, constants.MsgInvalidPhoneNumber, nil, fiber.StatusBadRequest)
	}

	// 6. Parse nomor telepon ke JID
	jid := utils.ParsePhoneNumber(req.PhoneNumber)

	// 7. Kirim pesan
	h.Logger.Info("Mengirim pesan personal", utils.Fields{
		"phone":  req.PhoneNumber,
		"length": len(req.Message),
	})

	sentTime, err := whatsClient.SendMessage(jid, req.Message)
	if err != nil {
		return h.SendError(c, fmt.Sprintf("%s: %v", constants.MsgSendFailure, err), err, fiber.StatusInternalServerError)
	}

	// 8. Log sukses
	h.LogSuccessResponse("Pesan personal berhasil dikirim", utils.Fields{
		"phone":     req.PhoneNumber,
		"timestamp": sentTime.Format(time.RFC3339),
	})

	// 9. Kirim respons sukses
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendSuccess,
		req.PhoneNumber,
		"personal",
		sentTime,
	))
}