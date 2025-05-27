package handler

import (
	"time"

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
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.MessageResponse{
			Success:   false,
			Message:   "Gagal mengirim pesan: WhatsApp sedang tidak terhubung",
			Timestamp: time.Now(), // Hanya timestamp respons yang disertakan
			// Tidak sertakan informasi koneksi sensitif lainnya
		})
	}

	var req model.PersonalMessageRequest

	// Validasi request
	if err := h.validatePersonalRequest(c, &req); err != nil {
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

// Metode validasi untuk request personal
func (h *MessageHandler) validatePersonalRequest(c *fiber.Ctx, req *model.PersonalMessageRequest) error {
	if err := c.BodyParser(req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, "Format request tidak valid", err, fiber.StatusBadRequest)
	}

	if req.PhoneNumber == "" || req.Message == "" {
		return h.SendError(c, "Nomor tujuan dan pesan notifikasi harus disediakan", nil, fiber.StatusBadRequest)
	}

	return nil
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
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.MessageResponse{
			Success:   false,
			Message:   "Gagal mengirim pesan: WhatsApp sedang tidak terhubung",
			Timestamp: time.Now(), // Hanya timestamp respons yang disertakan
			// Tidak sertakan informasi koneksi sensitif lainnya
		})
	}

	var req model.GroupMessageRequest

	// Validasi request
	if err := h.validateGroupRequest(c, &req); err != nil {
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

// Metode validasi untuk request grup
func (h *MessageHandler) validateGroupRequest(c *fiber.Ctx, req *model.GroupMessageRequest) error {
	if err := c.BodyParser(req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, "Format request tidak valid", err, fiber.StatusBadRequest)
	}

	if req.GroupID == "" || req.Message == "" {
		return h.SendError(c, "ID grup dan pesan notifikasi harus disediakan", nil, fiber.StatusBadRequest)
	}

	return nil
}
