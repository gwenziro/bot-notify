package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// DashboardController menangani halaman dashboard
type DashboardController struct {
	service *website.DashboardService
	logger  utils.LogrusEntry
}

// NewDashboardController membuat instance baru DashboardController
func NewDashboardController(service *website.DashboardService, logger utils.LogrusEntry) *DashboardController {
	return &DashboardController{
		service: service,
		logger:  logger.WithField("component", "dashboard-controller"),
	}
}

// DashboardPage menampilkan halaman dashboard utama
func (c *DashboardController) DashboardPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman dashboard")

	// Dapatkan data dari service
	dashData := c.service.GetDashboardData()

	// Render template dengan data
	return ctx.Render("dashboard", dashData)
}

// RefreshQRCode memuat ulang QR code untuk koneksi WhatsApp
func (c *DashboardController) RefreshQRCode(ctx *fiber.Ctx) error {
	c.logger.Info("Menerima permintaan refresh QR code")

	// Hubungkan ulang WhatsApp untuk mendapatkan QR code baru
	err := c.service.RefreshQRCode()
	if err != nil {
		c.logger.WithError(err).Error("Gagal menyambungkan untuk QR code baru")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memuat QR code baru: " + err.Error(),
		})
	}

	// Dapatkan status QR code
	qrStatus := c.service.GetQRCodeStatus()

	return ctx.JSON(fiber.Map{
		"success":     true,
		"message":     "QR code sedang dimuat",
		"qrAvailable": qrStatus.Available,
		"qrExpired":   qrStatus.Expired,
		"qrURL":       qrStatus.URL,
		"qrMessage":   qrStatus.Message,
	})
}

// DisconnectWhatsApp memutuskan koneksi WhatsApp
func (c *DashboardController) DisconnectWhatsApp(ctx *fiber.Ctx) error {
	c.logger.Info("Menerima permintaan disconnect WhatsApp")

	// Putuskan koneksi WhatsApp melalui service
	err := c.service.DisconnectWhatsApp()
	if err != nil {
		c.logger.WithError(err).Error("Gagal menghapus sesi WhatsApp")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus sesi: " + err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success":   true,
		"message":   "WhatsApp berhasil diputuskan",
		"connected": false,
	})
}
