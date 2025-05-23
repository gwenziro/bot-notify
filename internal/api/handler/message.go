package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// MessageHandler menangani endpoint pesan API
type MessageHandler struct {
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewMessageHandler membuat instance baru MessageHandler
func NewMessageHandler(whatsClient *client.Client) *MessageHandler {
	return &MessageHandler{
		whatsApp: whatsClient,
		logger:   utils.ForModule("handler-message"),
	}
}

// SendPersonal mengirim pesan ke nomor personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	var req model.PersonalMessageRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithError(err).Error("Gagal parsing request body")
		return c.Status(fiber.StatusBadRequest).JSON(
			model.NewErrorMessageResponse("Format request tidak valid", err, fiber.StatusBadRequest))
	}

	// Validasi input
	if req.PhoneNumber == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			model.NewErrorMessageResponse("Nomor tujuan dan pesan notifikasi harus disediakan", nil, fiber.StatusBadRequest))
	}

	// Konversi nomor telepon ke JID dan kirim pesan
	jid := client.ParsePhoneNumber(req.PhoneNumber)
	if err := h.whatsApp.SendMessage(jid, req.Message); err != nil {
		h.logger.WithFields(utils.Fields{
			"nomor": req.PhoneNumber,
			"error": err,
		}).Error("Gagal mengirim pesan personal")

		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewErrorMessageResponse("Gagal mengirim pesan", err, fiber.StatusInternalServerError))
	}

	h.logger.WithField("nomor", req.PhoneNumber).Info("Pesan personal berhasil dikirim")

	// Kirim response sukses
	return c.JSON(model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim!",
		jid.String(),
		"personal"))
}

// SendGroup mengirim pesan ke grup
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	var req model.GroupMessageRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithError(err).Error("Gagal parsing request body")
		return c.Status(fiber.StatusBadRequest).JSON(
			model.NewErrorMessageResponse("Format request tidak valid", err, fiber.StatusBadRequest))
	}

	// Validasi input
	if req.GroupID == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			model.NewErrorMessageResponse("ID grup dan pesan notifikasi harus disediakan", nil, fiber.StatusBadRequest))
	}

	// Konversi ID grup ke JID dan kirim pesan
	jid := client.ParseGroupID(req.GroupID)
	if err := h.whatsApp.SendMessage(jid, req.Message); err != nil {
		h.logger.WithFields(utils.Fields{
			"group": req.GroupID,
			"error": err,
		}).Error("Gagal mengirim pesan grup")

		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewErrorMessageResponse("Gagal mengirim pesan", err, fiber.StatusInternalServerError))
	}

	h.logger.WithField("group", req.GroupID).Info("Pesan grup berhasil dikirim")

	// Kirim response sukses
	return c.JSON(model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim ke grup!",
		jid.String(),
		"group"))
}
