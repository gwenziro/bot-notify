package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// DashboardController menangani halaman dashboard
type DashboardController struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewDashboardController membuat instance baru DashboardController
func NewDashboardController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DashboardController {
	return &DashboardController{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "dashboard-controller"),
	}
}

// DashboardPage menampilkan halaman dashboard utama
func (c *DashboardController) DashboardPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering halaman dashboard")

	// Persiapkan data untuk dashboard
	dashData := c.prepareDashboardData()

	// Render template dengan data
	return ctx.Render("dashboard", dashData)
}

// RefreshQRCode memuat ulang QR code untuk koneksi WhatsApp
func (c *DashboardController) RefreshQRCode(ctx *fiber.Ctx) error {
	c.logger.Info("Menerima permintaan refresh QR code")

	// Hubungkan ulang WhatsApp untuk mendapatkan QR code baru
	err := c.whatsApp.Connect()
	if err != nil {
		c.logger.WithError(err).Error("Gagal menyambungkan untuk QR code baru")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memuat QR code baru: " + err.Error(),
		})
	}

	// Tunggu sebentar agar QR code dibuat
	time.Sleep(1 * time.Second)

	// Dapatkan status QR code
	qrStatus := c.getQRCodeStatus()

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

	// Putuskan koneksi WhatsApp
	c.whatsApp.Disconnect()

	// Tunggu sejenak agar status koneksi diperbarui
	time.Sleep(300 * time.Millisecond)

	// Hapus sesi
	err := c.whatsApp.SessionManager.ClearSessions()
	if err != nil {
		c.logger.WithError(err).Error("Gagal menghapus sesi WhatsApp")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus sesi: " + err.Error(),
		})
	}

	// Verifikasi status koneksi
	state := c.whatsApp.GetConnectionState()

	return ctx.JSON(fiber.Map{
		"success":   true,
		"message":   "WhatsApp berhasil diputuskan",
		"connected": false,
		"status":    string(state.Status),
	})
}

// prepareDashboardData menyiapkan data untuk dashboard
func (c *DashboardController) prepareDashboardData() entity.DashboardData {
	// Buat data dasar
	baseURL := c.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	maskedToken := utils.MaskToken(c.config.Auth.AccessToken)
	data := entity.NewDashboardData(baseURL, maskedToken)

	// Dapatkan status koneksi WhatsApp
	connectionState, _ := c.whatsApp.GetConnectionStateSafe()
	data.IsConnected = connectionState.IsConnected
	data.ConnectionStatus = string(connectionState.Status)
	data.LastActivity = connectionState.LastActivity
	data.LastActivityFormatted = utils.FormatTimeIndonesia(&connectionState.LastActivity)

	// Jika terhubung, tambahkan informasi koneksi
	if data.IsConnected {
		// Dapatkan profil info seperti di profile_handler.go
		deviceInfo := c.whatsApp.GetDeviceInfo()

		// Periksa status koneksi dari WhatsApp client langsung
		connectionInfo := c.whatsApp.GetConnectionInfo()

		// Dapatkan waktu connectedSince yang lebih akurat dari profil WhatsApp
		// Ini akan lebih konsisten dan tidak bergantung pada LastActivity
		if connTime, ok := connectionInfo["connected_since"].(time.Time); ok && !connTime.IsZero() {
			// Gunakan waktu koneksi dari info koneksi
			data.ConnectedSince = connTime
			data.ConnectedSinceFormatted = utils.FormatTimeIndonesia(&connTime)

			// Hitung durasi koneksi berdasarkan waktu yang akurat
			duration := time.Since(connTime)
			// Batasi durasi maksimal untuk mencegah nilai yang tidak masuk akal
			if duration > 365*24*time.Hour {
				duration = 24 * time.Hour // Default 1 hari jika berlebihan
			}
			data.ConnectionDuration = utils.FormatUptime(duration)

			c.logger.Debug("Informasi durasi koneksi dari GetConnectionInfo", utils.Fields{
				"connected_since": connTime.Format(time.RFC3339),
				"duration":        data.ConnectionDuration,
			})
		} else {
			// Fallback ke metode lama jika metode baru gagal
			if !connectionState.ConnectedSince.IsZero() {
				data.ConnectedSince = connectionState.ConnectedSince
				data.ConnectedSinceFormatted = utils.FormatTimeIndonesia(&connectionState.ConnectedSince)

				duration := time.Since(connectionState.ConnectedSince)
				if duration > 365*24*time.Hour {
					duration = 24 * time.Hour
				}
				data.ConnectionDuration = utils.FormatUptime(duration)

				c.logger.Debug("Fallback ke ConnectionState untuk durasi koneksi", utils.Fields{
					"connected_since": connectionState.ConnectedSince.Format(time.RFC3339),
					"duration":        data.ConnectionDuration,
				})
			} else {
				// Jika masih tidak ada, gunakan LastActivity sebagai pilihan terakhir
				c.logger.Warn("Tidak ada ConnectedSince yang valid, menggunakan LastActivity", utils.Fields{
					"last_activity": connectionState.LastActivity.Format(time.RFC3339),
				})

				data.ConnectedSince = connectionState.LastActivity
				data.ConnectedSinceFormatted = utils.FormatTimeIndonesia(&connectionState.LastActivity)
				data.ConnectionDuration = "Tidak diketahui"
			}
		}

		// Tambahkan informasi profil - pastikan semua data diisi
		// Set default values untuk mencegah nil
		data.PhoneNumber = "Tidak tersedia"
		data.ContactName = "Tidak tersedia"

		if jid, ok := deviceInfo["id"].(string); ok && jid != "" {
			data.PhoneNumber = client.FormatWhatsAppNumber(jid)
		}

		if pushName, ok := deviceInfo["push_name"].(string); ok && pushName != "" {
			data.ContactName = pushName
		} else {
			// Gunakan nomor telepon sebagai fallback jika tidak ada nama
			data.ContactName = data.PhoneNumber
		}

		// Perbaikan URL foto profil
		if pictureURL, ok := deviceInfo["picture_url"].(string); ok && pictureURL != "" {
			// Pastikan URL menggunakan HTTPS
			if strings.HasPrefix(pictureURL, "http:") {
				pictureURL = "https:" + strings.TrimPrefix(pictureURL, "http:")
			}
			data.ProfilePictureURL = pictureURL
			c.logger.Debug("Menggunakan foto profil", utils.Fields{
				"url": pictureURL,
			})
		} else {
			c.logger.Debug("Foto profil tidak tersedia")
		}

		// Dapatkan jumlah pesan terkirim dari WhatsApp client
		messagesSent := c.whatsApp.GetMessagesSent()
		data.MessagesSent = int(messagesSent)

		c.logger.Debug("Retrieved message statistics", utils.Fields{
			"messages_sent": messagesSent,
		})
	} else {
		// Jika tidak terhubung, dapatkan informasi QR code
		qrStatus := c.getQRCodeStatus()
		data.QRCodeAvailable = qrStatus.Available
		data.QRCodeExpired = qrStatus.Expired
		data.QRCodeURL = qrStatus.URL
		data.QRCodeMessage = qrStatus.Message
	}

	return data
}

// QRStatus adalah struktur data untuk status QR code
type QRStatus struct {
	Available bool
	Expired   bool
	URL       string
	Message   string
}

// getQRCodeStatus mendapatkan status QR code
func (c *DashboardController) getQRCodeStatus() QRStatus {
	// Periksa apakah QR handler tersedia
	qrHandler := c.whatsApp.SessionManager.GetQRHandler()
	if qrHandler == nil {
		return QRStatus{
			Available: false,
			Expired:   false,
			URL:       "",
			Message:   "QR code handler tidak tersedia. Silakan coba refresh.",
		}
	}

	// Dapatkan timestamp QR code
	timestamp := qrHandler.GetQRCodeTimestamp()
	qrExists := !timestamp.IsZero()

	// QR dianggap kedaluwarsa jika berusia lebih dari 25 detik
	maxAgeMins := 25.0 / 60.0 // 25 detik dalam menit
	expired := qrExists && qrHandler.IsQRCodeExpired(maxAgeMins)

	// QR dianggap tersedia jika QR ada DAN tidak kedaluwarsa
	available := qrExists && !expired

	// Tentukan pesan berdasarkan status
	var message string
	var qrURL string

	if expired {
		message = "QR code sudah kedaluwarsa. Silakan klik tombol 'Segarkan QR' untuk mendapatkan QR code baru."
	} else if !available {
		message = "QR code belum tersedia. Silakan klik tombol 'Segarkan QR' untuk mendapatkan QR code baru."
	} else {
		message = "Silakan pindai kode QR berikut dengan WhatsApp di ponsel Anda untuk menghubungkan Bot:"
		qrURL = fmt.Sprintf("%s/api/qr/image?t=%d", c.config.Server.BaseURL, timestamp.Unix())
	}

	return QRStatus{
		Available: available,
		Expired:   expired,
		URL:       qrURL,
		Message:   message,
	}
}
