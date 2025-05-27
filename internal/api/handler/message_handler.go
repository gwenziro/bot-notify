package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
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

// SendPersonal mengirim pesan ke nomor personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// Dapatkan status koneksi terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas
	if !state.IsConnected {
		h.Logger.Info("Permintaan kirim pesan personal saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
	}

	var req model.PersonalMessageRequest

	// Validasi request
	if err := h.ValidateRequest(c, &req, map[string]func() string{
		"phoneNumber": func() string { return req.PhoneNumber },
		"message":     func() string { return req.Message },
	}); err != nil {
		return err
	}

	// Kirim pesan
	jid := client.ParsePhoneNumber(req.PhoneNumber)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan personal")
		return h.SendError(c, "Gagal mengirim pesan", err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan personal berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim!",
		jid.String(),
		"personal"))
}

// SendGroup mengirim pesan ke grup
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// Dapatkan status koneksi terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas
	if !state.IsConnected {
		h.Logger.Info("Permintaan kirim pesan grup saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
	}

	var req model.GroupMessageRequest

	// Validasi request
	if err := h.ValidateRequest(c, &req, map[string]func() string{
		"groupID": func() string { return req.GroupID },
		"message": func() string { return req.Message },
	}); err != nil {
		return err
	}

	// Kirim pesan
	jid := client.ParseGroupID(req.GroupID)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan grup")
		return h.SendError(c, "Gagal mengirim pesan", err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan grup berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim ke grup!",
		jid.String(),
		"group"))
}
