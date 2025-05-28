package handler

import (
	"fmt"

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

// SendError mengirim respons error standar
func (h *BaseHandler) SendError(c *fiber.Ctx, message string, err error, code int) error {
	return c.Status(code).JSON(model.NewErrorMessageResponse(message, err, code))
}

// SendSuccess mengirim respons sukses standar dengan data
func (h *BaseHandler) SendSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(data)
}

// SendDisconnectedResponse mengirim respons standar untuk kondisi tidak terhubung
// Return: false agar bisa digunakan sebagai return value dalam if statement
func (h *BaseHandler) SendDisconnectedResponse(c *fiber.Ctx, customMessage string) bool {
	if customMessage == "" {
		customMessage = "WhatsApp sedang tidak terhubung"
	}

	// Gunakan HTTP 503 Service Unavailable untuk konsistensi
	c.Status(fiber.StatusServiceUnavailable).JSON(model.NewBaseResponse(
		false,
		customMessage,
	))

	return false
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

// GetConnectionState mengambil state koneksi dengan penanganan error yang tepat
func (h *BaseHandler) GetConnectionState() (client.ConnectionState, error) {
	if h.WhatsApp == nil {
		h.Logger.Error("WhatsApp client is nil")
		return client.ConnectionState{}, fiber.NewError(fiber.StatusInternalServerError, "WhatsApp client not initialized")
	}

	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		h.Logger.WithError(err).Error("Failed to get connection state")
		return client.ConnectionState{}, fiber.NewError(fiber.StatusInternalServerError, "Failed to get connection state")
	}

	return state, nil
}

// CheckWhatsAppConnection memeriksa dan mengirimkan respons disconnect jika perlu
// Return: true jika terhubung, false jika tidak terhubung (dan respons sudah dikirim)
func (h *BaseHandler) CheckWhatsAppConnection(c *fiber.Ctx, customMessage string) bool {
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
		return false
	}

	if !state.IsConnected {
		h.Logger.Info(fmt.Sprintf("Permintaan endpoint %s saat WhatsApp tidak terhubung", c.Path()))
		return h.SendDisconnectedResponse(c, customMessage)
	}

	return true
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

// ValidateRequest memvalidasi request body secara generik
func (h *BaseHandler) ValidateRequest(c *fiber.Ctx, req interface{}, requiredFields map[string]func() string) error {
	// Parse request body
	if err := c.BodyParser(req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, "Format request tidak valid", err, fiber.StatusBadRequest)
	}

	// Validasi required fields
	for field, getValue := range requiredFields {
		if getValue() == "" {
			return h.SendError(c, fmt.Sprintf("Field %s tidak boleh kosong", field), nil, fiber.StatusBadRequest)
		}
	}

	return nil
}
