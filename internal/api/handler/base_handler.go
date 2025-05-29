package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
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

// CheckWhatsAppConnection memeriksa dan mengirimkan respons disconnect jika perlu
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

// ValidateRequest memvalidasi request body secara generik
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

// ValidateInput memvalidasi input request dengan menggabungkan parsing dan validasi
func (h *BaseHandler) ValidateInput(c *fiber.Ctx, req interface{}, validator func(interface{}) error) error {
	if err := c.BodyParser(req); err != nil {
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	if validator != nil {
		if err := validator(req); err != nil {
			return h.SendError(c, err.Error(), nil, fiber.StatusBadRequest)
		}
	}

	return nil
}

// HandleError menangani error dengan logging dan respons yang konsisten
func (h *BaseHandler) HandleError(c *fiber.Ctx, message string, err error, statusCode int) error {
	logFields := utils.Fields{"path": c.Path()}

	if err != nil {
		logFields["error"] = err.Error()
		h.Logger.WithFields(logFields).Error(message)
	} else {
		h.Logger.WithFields(logFields).Warn(message)
	}

	return c.Status(statusCode).JSON(model.NewBaseErrorResponse(message, err, statusCode))
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
