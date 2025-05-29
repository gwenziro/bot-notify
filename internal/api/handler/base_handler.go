package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// BaseHandler berisi fungsionalitas umum untuk semua handler API
type BaseHandler struct {
	WhatsApp *client.Client    // Klien WhatsApp untuk operasi pesan
	Logger   utils.LogrusEntry // Logger untuk mencatat aktivitas
}

// NewBaseHandler membuat instance baru BaseHandler
func NewBaseHandler(whatsClient *client.Client, module string) BaseHandler {
	return BaseHandler{
		WhatsApp: whatsClient,
		Logger:   utils.ForModule(module),
	}
}

// SendError mengirim respons error standar dengan logging otomatis
func (h *BaseHandler) SendError(c *fiber.Ctx, message string, err error, code int) error {
	if err != nil {
		h.Logger.WithError(err).Error(message)
	} else {
		h.Logger.Error(message)
	}
	return c.Status(code).JSON(model.NewBaseErrorResponse(message, err, code))
}

// SendSuccess mengirim respons sukses standar dengan data
func (h *BaseHandler) SendSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(data)
}

// SendDisconnectedResponse mengirim respons standar untuk kondisi tidak terhubung
func (h *BaseHandler) SendDisconnectedResponse(c *fiber.Ctx, customMessage string) bool {
	if customMessage == "" {
		customMessage = constants.MsgNotConnected
	}

	// Gunakan HTTP 503 Service Unavailable untuk konsistensi
	c.Status(fiber.StatusServiceUnavailable).JSON(model.NewBaseResponse(false, customMessage))

	return false
}

// CheckWhatsAppConnection memeriksa koneksi WhatsApp dan mengirimkan respons jika tidak terhubung
func (h *BaseHandler) CheckWhatsAppConnection(c *fiber.Ctx, customMessage string) bool {
	// Validasi client tidak nil
	if h.WhatsApp == nil {
		h.SendError(c, constants.MsgClientNotAvailable, nil, fiber.StatusInternalServerError)
		return false
	}

	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		h.SendError(c, fmt.Sprintf(constants.MsgStatusFailed, err), nil, fiber.StatusInternalServerError)
		return false
	}

	if !state.IsConnected {
		h.Logger.Info(fmt.Sprintf("Permintaan endpoint %s saat WhatsApp tidak terhubung", c.Path()))
		return h.SendDisconnectedResponse(c, customMessage)
	}

	return true
}

// ValidateRequest memvalidasi request body dan field yang diperlukan
func (h *BaseHandler) ValidateRequest(c *fiber.Ctx, req interface{}, requiredFields map[string]func() string) error {
	// Parse request body
	if err := c.BodyParser(req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	// Validasi required fields
	for field, getValue := range requiredFields {
		if getValue() == "" {
			return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, field), nil, fiber.StatusBadRequest)
		}
	}

	return nil
}

// ParseAndValidateBody melakukan parsing dan validasi request body
func (h *BaseHandler) ParseAndValidateBody(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}
	return nil
}

// ValidateRequiredField memvalidasi field wajib dengan error handling yang konsisten
func (h *BaseHandler) ValidateRequiredField(c *fiber.Ctx, fieldName string, fieldValue string) error {
	if fieldValue == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, fieldName), nil, fiber.StatusBadRequest)
	}
	return nil
}

// LogDebugRequest mencatat informasi debug tentang request
func (h *BaseHandler) LogDebugRequest(c *fiber.Ctx, handlerName string) {
	h.Logger.Debug(fmt.Sprintf("%s dipanggil", handlerName), utils.Fields{
		"path":   c.Path(),
		"method": c.Method(),
		"ip":     c.IP(),
	})
}

// LogSuccessResponse mencatat informasi tentang respons sukses
func (h *BaseHandler) LogSuccessResponse(message string, fields utils.Fields) {
	h.Logger.WithFields(fields).Info(message)
}

// FormatConnectedSince memformat waktu koneksi dalam format Indonesia
func (h *BaseHandler) FormatConnectedSince(state client.ConnectionState) string {
	if !state.IsConnected {
		return ""
	}

	// Pilih waktu yang valid
	timeToFormat := state.ConnectedSince
	if timeToFormat.IsZero() {
		timeToFormat = state.LastActivity
	}

	return utils.FormatTimeIndonesia(&timeToFormat)
}
