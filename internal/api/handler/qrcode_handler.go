package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/session"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// QR code expiration time constant
const QrCodeExpirationMinutes = 25.0 / 60.0 // 25 detik dinyatakan dalam menit

// QRCodeHandler menangani endpoint QR code untuk koneksi WhatsApp
type QRCodeHandler struct {
	BaseHandler
	sessionMgr *session.Manager // Manager untuk sesi WhatsApp
	maxAgeMins float64          // Masa berlaku maksimum QR code dalam menit
}

// NewQRCodeHandler membuat instance baru QRCodeHandler
func NewQRCodeHandler(whatsClient *client.Client) *QRCodeHandler {
	return &QRCodeHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-qrcode"),
		sessionMgr:  whatsClient.SessionManager,
		maxAgeMins:  QrCodeExpirationMinutes,
	}
}

// GetStatus mengembalikan status QR code saat ini
// Endpoint: GET /api/qr/status
func (h *QRCodeHandler) GetStatus(c *fiber.Ctx) error {
	// Log debug request
	h.LogDebugRequest(c, "GetStatus")

	// Periksa status koneksi WhatsApp terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika WhatsApp sudah terhubung, berikan respons khusus
	if state.IsConnected {
		// WhatsApp sudah terhubung, tidak perlu QR code
		response := model.NewQRCodeStatusResponse(false, false, constants.MsgQrConnected)
		response.ConnectedStatus = true

		h.LogSuccessResponse("QR code tidak tersedia karena sudah terhubung", utils.Fields{
			"connected": true,
		})

		return c.Status(fiber.StatusOK).JSON(response)
	}

	// Jika WhatsApp tidak terhubung, periksa status QR code
	qrHandler := h.WhatsApp.SessionManager.GetQRHandler()
	if qrHandler == nil {
		return h.SendError(c, "QR handler tidak tersedia", nil, fiber.StatusInternalServerError)
	}

	// Dapatkan timestamp QR code
	timestamp := qrHandler.GetQRCodeTimestamp()
	qrExists := !timestamp.IsZero()
	expired := qrExists && qrHandler.IsQRCodeExpired(h.maxAgeMins)

	// QR dianggap tersedia jika QR ada DAN tidak kedaluwarsa
	available := qrExists && !expired

	// Tentukan pesan berdasarkan status
	var message string
	if expired {
		message = constants.MsgQrExpired
	} else if !available {
		message = constants.MsgQrNotAvailable
	} else {
		message = constants.MsgQrAvailable
	}

	// Siapkan response sesuai status
	response := model.NewQRCodeStatusResponse(available, expired, message)

	// PERUBAHAN: Selalu gunakan StatusOK (200) tanpa memandang status QR code
	statusCode := fiber.StatusOK

	h.LogSuccessResponse("Status QR code berhasil diambil", utils.Fields{
		"available": available,
		"expired":   expired,
		"status":    statusCode,
	})

	return c.Status(statusCode).JSON(response)
}

// GetImage mengembalikan gambar QR code
// Endpoint: GET /api/qr/image
func (h *QRCodeHandler) GetImage(c *fiber.Ctx) error {
	// Log debug request
	h.LogDebugRequest(c, "GetImage")

	// Periksa ketersediaan QR code berdasarkan status koneksi
	canContinue, err := h.checkQRAvailability(c)
	if !canContinue {
		// Untuk endpoint gambar, berikan respons image-friendly
		if c.Get("Accept") == "application/json" {
			return err
		}

		// Untuk browser atau permintaan non-JSON, kirim respons 404 yang sederhana
		h.Logger.Warn("QR code tidak tersedia", utils.Fields{
			"path": c.Path(),
		})

		return c.Status(fiber.StatusNotFound).SendString("QR code tidak tersedia")
	}

	// Dapatkan QR handler dari session manager
	qrHandler := h.sessionMgr.GetQRHandler()
	if qrHandler == nil {
		h.Logger.Error("QR handler tidak tersedia")
		return c.Status(fiber.StatusInternalServerError).SendString("QR handler tidak tersedia")
	}

	// Dapatkan path QR code
	qrPath := qrHandler.GetQRCodePath()

	// Jika QR code kedaluwarsa atau tidak ada, return 404
	if qrHandler.IsQRCodeExpired(h.maxAgeMins) {
		h.Logger.Warn("QR code kedaluwarsa", utils.Fields{
			"age_mins": h.maxAgeMins,
		})

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "QR code kedaluwarsa atau tidak tersedia",
			"code":    fiber.StatusNotFound,
		})
	}

	h.LogSuccessResponse("QR code image berhasil diambil", utils.Fields{
		"path": qrPath,
	})

	// Send QR code sebagai file
	return c.SendFile(qrPath)
}

// checkQRAvailability memeriksa apakah QR code tersedia berdasarkan status koneksi
func (h *QRCodeHandler) checkQRAvailability(c *fiber.Ctx) (canContinue bool, err error) {
	// Dapatkan status koneksi
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return false, h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika sudah terhubung, QR code tidak relevan
	if state.IsConnected || h.WhatsApp.IsLoggedIn() {
		return false, h.SendSuccess(c, model.NewQRCodeStatusResponse(false, false, constants.MsgQrConnected))
	}

	// Jika tidak terhubung tetapi belum menjalankan reconnect
	if !state.IsConnected && state.Status != client.StatusConnecting {
		return false, h.SendSuccess(c, model.NewQRCodeStatusResponse(false, false, constants.MsgQrNotConnected))
	}

	return true, nil
}
