package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/session"
)

// QRCodeHandler menangani endpoint QR code API
type QRCodeHandler struct {
	BaseHandler
	sessionMgr *session.Manager
	maxAgeMins int
}

// NewQRCodeHandler membuat instance baru QRCodeHandler
func NewQRCodeHandler(whatsClient *client.Client) *QRCodeHandler {
	return &QRCodeHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-qrcode"),
		sessionMgr:  whatsClient.SessionManager,
		maxAgeMins:  5, // QR code kedaluwarsa setelah 5 menit
	}
}

// GetStatus mengembalikan status QR code
func (h *QRCodeHandler) GetStatus(c *fiber.Ctx) error {
	// Dapatkan status koneksi
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika sudah terhubung, QR code tidak relevan
	if state.IsConnected || h.WhatsApp.IsLoggedIn() {
		return h.SendSuccess(c, model.QRCodeStatusResponse{
			Available: false,
			Expired:   false,
			Message:   "QR code tidak diperlukan karena WhatsApp sudah terhubung",
			Timestamp: nil,
		})
	}

	// Jika tidak terhubung tetapi belum menjalankan reconnect
	if !state.IsConnected && state.Status != client.StatusConnecting {
		return h.SendSuccess(c, model.QRCodeStatusResponse{
			Available: false,
			Expired:   false,
			Message:   "QR code belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu",
			Timestamp: nil,
		})
	}

	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		return h.SendError(c, "QR handler tidak tersedia", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan data QR code
	data, timestamp := qrHandler.GetQRCodeData()
	hasQR := data != ""
	isExpired := qrHandler.IsQRCodeExpired(h.maxAgeMins)

	return h.SendSuccess(c, model.NewQRCodeStatusResponse(hasQR && !isExpired, isExpired, timestamp))
}

// GetImage mengembalikan gambar QR code
func (h *QRCodeHandler) GetImage(c *fiber.Ctx) error {
	// Dapatkan status koneksi
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika sudah terhubung, QR code tidak relevan
	if state.IsConnected || h.WhatsApp.IsLoggedIn() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "QR code tidak diperlukan karena WhatsApp sudah terhubung",
			"code":    fiber.StatusNotFound,
		})
	}

	// Jika tidak terhubung tetapi belum menjalankan reconnect
	if !state.IsConnected && state.Status != client.StatusConnecting {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Gambar QR belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu untuk mendapatkan gambar QR",
			"code":    fiber.StatusNotFound,
		})
	}

	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("QR handler tidak tersedia")
	}

	// Dapatkan path QR code
	qrPath := qrHandler.GetQRCodePath()

	// Jika QR code kedaluwarsa atau tidak ada, return 404
	if qrHandler.IsQRCodeExpired(h.maxAgeMins) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "QR code kedaluwarsa atau tidak tersedia",
			"code":    fiber.StatusNotFound,
		})
	}
	if qrHandler.IsQRCodeExpired(h.maxAgeMins) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "QR code kedaluwarsa atau tidak tersedia",
			"code":    fiber.StatusNotFound,
		})
	}

	// Send QR code sebagai file
	return c.SendFile(qrPath)
}
