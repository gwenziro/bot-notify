package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/website"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/model"
)

// DashboardController menangani halaman dashboard WhatsApp
type DashboardController struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	statsService *website.StatisticsService
	startTime    time.Time
}

// NewDashboardController membuat instance baru DashboardController
func NewDashboardController(
	cfg *config.Config,
	whatsClient *client.Client,
	statsService *website.StatisticsService,
	logger utils.LogrusEntry,
) *DashboardController {
	return &DashboardController{
		config:       cfg,
		whatsApp:     whatsClient,
		logger:       logger.WithField("component", "dashboard-controller"),
		statsService: statsService,
		startTime:    time.Now(),
	}
}

// DashboardPage menampilkan halaman dashboard utama
func (c *DashboardController) DashboardPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering dashboard page")

	// 1. Dapatkan status koneksi WhatsApp
	connectionState := c.whatsApp.GetConnectionState()

	// 2. Hitung uptime aplikasi (waktu sejak aplikasi dimulai)
	uptime := utils.FormatUptime(time.Since(c.startTime))

	// 3. Persiapkan model dasar
	dashboardModel := model.NewDashboardModel(connectionState, uptime)

	// 4. Dapatkan jumlah pesan terkirim dari service, bukan dari controller
	dashboardModel.MessagesSent = c.statsService.GetMessageCount()

	// 5. Persiapkan data untuk template
	baseURL := c.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	dashboardModel.BaseURL = baseURL

	// Tambahkan token akses untuk API calls dari JavaScript
	dashboardModel.APIToken = c.config.Auth.AccessToken
	dashboardModel.MaskedToken = utils.MaskToken(c.config.Auth.AccessToken)

	// 6. Jika terhubung, dapatkan statistik detail
	if connectionState.IsConnected {
		// 6.1 Data "Terhubung Sejak"
		// Gunakan ConnectedSince dari connectionState jika tersedia, atau LastActivity jika tidak
		connectTime := connectionState.ConnectedSince
		if connectTime.IsZero() {
			connectTime = connectionState.LastActivity
		}
		dashboardModel.ConnectedSince = connectTime
		dashboardModel.ConnectedSinceFormatted = utils.FormatTimeShort(&connectTime)

		// 6.2 Data "Terhubung Selama"
		// Hitung durasi dari waktu terhubung hingga sekarang
		connectedDuration := time.Since(connectTime)
		dashboardModel.ConnectedDuration = utils.FormatUptime(connectedDuration)

		// 6.3 Data "Aktivitas Terakhir"
		dashboardModel.LastActivity = connectionState.LastActivity
		dashboardModel.LastActivityFormatted = utils.FormatTimeShort(&connectionState.LastActivity)

		// 6.4 Data kontak WhatsApp
		selfID := c.whatsApp.GetSelfID()
		if selfID != nil {
			// Format JID menjadi nomor telepon yang lebih mudah dibaca
			dashboardModel.PhoneNumber = utils.FormatWhatsAppNumber(selfID.String())

			// Gunakan fungsi GetContactName untuk mendapatkan nama kontak
			dashboardModel.ContactName = c.whatsApp.GetContactName(*selfID, dashboardModel.PhoneNumber)
		}

		// Log informasi yang berhasil diambil
		c.logger.Debug("Dashboard statistics retrieved", utils.Fields{
			"app_uptime":         uptime,
			"connected_since":    dashboardModel.ConnectedSinceFormatted,
			"connected_duration": dashboardModel.ConnectedDuration,
			"last_activity":      dashboardModel.LastActivityFormatted,
			"messages_sent":      dashboardModel.MessagesSent,
		})
	} else {
		// 7. Jika tidak terhubung, persiapkan QR code
		c.prepareQRCodeInfo(&dashboardModel, connectionState)
	}

	// 8. Render dashboard dengan model
	return ctx.Render("dashboard", dashboardModel)
}

// prepareQRCodeInfo menyiapkan informasi QR code untuk dashboard
func (c *DashboardController) prepareQRCodeInfo(dashboardModel *model.DashboardModel, connectionState client.ConnectionState) {
	// Default - QR code tidak tersedia
	dashboardModel.QRCodeAvailable = false
	dashboardModel.QRCodeExpired = false
	dashboardModel.QRCodePath = ""
	dashboardModel.ShowReconnectButton = true

	// Jika WhatsApp sudah terhubung, tidak perlu QR code
	if connectionState.IsConnected {
		dashboardModel.QRCodeMessage = "WhatsApp sudah terhubung"
		dashboardModel.ShowReconnectButton = false
		return
	}

	// Dapatkan QR handler dari session manager
	qrHandler := c.whatsApp.SessionManager.GetQRHandler()
	if qrHandler == nil {
		dashboardModel.QRCodeMessage = "QR handler tidak tersedia"
		return
	}

	// Ambil status QR code dari handler
	available, expired, path, message, showReconnect := qrHandler.GetQRCodeStatus()

	dashboardModel.QRCodeAvailable = available
	dashboardModel.QRCodeExpired = expired
	dashboardModel.QRCodePath = path
	dashboardModel.QRCodeMessage = message
	dashboardModel.ShowReconnectButton = showReconnect

	// Log status QR code
	c.logger.Debug("QR code status for dashboard", utils.Fields{
		"available":      available,
		"expired":        expired,
		"has_path":       path != "",
		"show_reconnect": showReconnect,
	})
}
