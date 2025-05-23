package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// QRCodeController menangani halaman QR code
type QRCodeController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewQRCodeController membuat instance baru QRCodeController
func NewQRCodeController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *QRCodeController {
	return &QRCodeController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "qrcode-controller"),
	}
}

// QRCodePage menampilkan QR code untuk scanning
func (c *QRCodeController) QRCodePage(ctx *fiber.Ctx) error {
	return ctx.Render("qrcode", fiber.Map{
		"Title":            "Scan QR Code",
		"QRCodeEndpoint":   "/api/qr/image",
		"QRStatusEndpoint": "/api/qr/status",
	})
}
