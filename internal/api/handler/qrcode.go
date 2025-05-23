package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/session"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// QRCodeHandler menangani endpoint QR code API
type QRCodeHandler struct {
	whatsApp   *client.Client
	sessionMgr *session.Manager
	logger     utils.LogrusEntry
	maxAgeMins int
}

// NewQRCodeHandler membuat instance baru QRCodeHandler
func NewQRCodeHandler(whatsClient *client.Client) *QRCodeHandler {
	return &QRCodeHandler{
		whatsApp:   whatsClient,
		sessionMgr: whatsClient.SessionManager,
		logger:     utils.ForModule("handler-qrcode"),
		maxAgeMins: 5, // QR code kedaluwarsa setelah 5 menit
	}
}

// GetStatus mengembalikan status QR code
func (h *QRCodeHandler) GetStatus(c *fiber.Ctx) error {
	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"sukses":    false,
			"pesan":     "QR handler tidak tersedia",
			"available": false,
		})
	}

	// Dapatkan data QR code
	data, timestamp := qrHandler.GetQRCodeData()
	hasQR := data != ""
	isExpired := qrHandler.IsQRCodeExpired(h.maxAgeMins)

	return c.JSON(fiber.Map{
		"sukses":    true,
		"available": hasQR && !isExpired,
		"expired":   isExpired,
		"timestamp": timestamp,
	})
}

// GetImage mengembalikan gambar QR code
func (h *QRCodeHandler) GetImage(c *fiber.Ctx) error {
	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("QR handler tidak tersedia")
	}

	// Dapatkan path QR code
	qrPath := qrHandler.GetQRCodePath()

	// Jika QR code kedaluwarsa atau tidak ada, return 404
	if qrHandler.IsQRCodeExpired(h.maxAgeMins) {
		return c.Status(fiber.StatusNotFound).SendString("QR code kedaluwarsa atau tidak tersedia")
	}

	// Send QR code sebagai file
	return c.SendFile(qrPath)
}
