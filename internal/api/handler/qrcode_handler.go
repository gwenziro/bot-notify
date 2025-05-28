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

// GetStatus mengembalikan status QR code saat ini
func (h *QRCodeHandler) GetStatus(c *fiber.Ctx) error {
	// Periksa status koneksi WhatsApp terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika WhatsApp sudah terhubung, berikan respons khusus
	if state.IsConnected {
		// WhatsApp sudah terhubung, tidak perlu QR code
		response := model.NewQRCodeStatusResponse(false, false)

		// Override pesan untuk kasus sudah terhubung
		response.Message = "QR code tidak tersedia: WhatsApp sudah terhubung"
		response.ConnectedStatus = true // Tambahkan informasi bahwa sudah terhubung

		return c.Status(fiber.StatusOK).JSON(response)
	}

	// Tentukan status code yang sesuai
	statusCode := fiber.StatusOK

	// Jika WhatsApp tidak terhubung, maka QR code tidak tersedia
	if !state.IsConnected {
		statusCode = fiber.StatusServiceUnavailable
	}

	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		return h.SendError(c, "QR handler tidak tersedia", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan data QR code
	data, _ := qrHandler.GetQRCodeData()
	hasQR := data != ""
	isExpired := qrHandler.IsQRCodeExpired(h.maxAgeMins)

	response := model.NewQRCodeStatusResponse(hasQR && !isExpired, isExpired)

	// Return dengan status code yang sesuai
	return c.Status(statusCode).JSON(response)
}

// checkQRAvailability memeriksa apakah QR code tersedia berdasarkan status koneksi
func (h *QRCodeHandler) checkQRAvailability(c *fiber.Ctx) (canContinue bool, err error) {
	// Dapatkan status koneksi
	state, err := h.GetConnectionState()
	if err != nil {
		return false, h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika sudah terhubung, QR code tidak relevan
	if state.IsConnected || h.WhatsApp.IsLoggedIn() {
		return false, h.SendSuccess(c, model.NewQRCodeStatusResponse(false, false))
	}

	// Jika tidak terhubung tetapi belum menjalankan reconnect
	if !state.IsConnected && state.Status != client.StatusConnecting {
		return false, h.SendSuccess(c, model.NewQRCodeStatusResponse(false, false))
	}

	return true, nil
}

// GetImage mengembalikan gambar QR code
func (h *QRCodeHandler) GetImage(c *fiber.Ctx) error {
	// Periksa ketersediaan QR code berdasarkan status koneksi
	canContinue, err := h.checkQRAvailability(c)
	if !canContinue {
		// Untuk endpoint gambar, berikan respons image-friendly
		if c.Get("Accept") == "application/json" {
			return err
		}

		// Untuk browser atau permintaan non-JSON, kirim respons 404 yang sederhana
		return c.Status(fiber.StatusNotFound).SendString("QR code tidak tersedia")
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

	// Send QR code sebagai file
	return c.SendFile(qrPath)
}
