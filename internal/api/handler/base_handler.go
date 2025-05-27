package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// BaseHandler berisi fungsionalitas umum untuk semua handler
type BaseHandler struct {
	WhatsApp *client.Client
	Logger   utils.LogrusEntry
}

// NewBaseHandler membuat instance BaseHandler baru
func NewBaseHandler(whatsClient *client.Client, module string) BaseHandler {
	return BaseHandler{
		WhatsApp: whatsClient,
		Logger:   utils.ForModule(module),
	}
}

// CheckConnection memeriksa apakah WhatsApp terhubung
func (h *BaseHandler) CheckConnection(c *fiber.Ctx) error {
	if h.WhatsApp == nil {
		h.Logger.Error("WhatsApp client is nil")
		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewErrorMessageResponse("Server error: WhatsApp client not initialized", nil, fiber.StatusInternalServerError))
	}

	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		h.Logger.WithError(err).Error("Failed to get connection state")
		return c.Status(fiber.StatusInternalServerError).JSON(
			model.NewErrorMessageResponse("Server error: Failed to get connection state", err, fiber.StatusInternalServerError))
	}

	if !state.IsConnected {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			model.NewErrorMessageResponse("WhatsApp tidak terhubung", nil, fiber.StatusServiceUnavailable))
	}

	return nil
}

// SendError mengirim respons error standar
func (h *BaseHandler) SendError(c *fiber.Ctx, message string, err error, code int) error {
	return c.Status(code).JSON(model.NewErrorMessageResponse(message, err, code))
}

// SendSuccess mengirim respons sukses standar dengan data
func (h *BaseHandler) SendSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(data)
}
