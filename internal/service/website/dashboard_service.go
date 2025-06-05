package website

import (
	"fmt"
	"strings"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// DashboardService menyediakan fungsionalitas untuk halaman dashboard
type DashboardService struct {
	config   *config.Config
	whatsApp *client.Client
	logger   utils.LogrusEntry
}

// NewDashboardService membuat instance service dashboard baru
func NewDashboardService(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DashboardService {
	return &DashboardService{
		config:   cfg,
		whatsApp: whatsClient,
		logger:   logger.WithField("component", "dashboard-service"),
	}
}

// GetDashboardData menyiapkan semua data yang diperlukan untuk dashboard
func (s *DashboardService) GetDashboardData() entity.DashboardData {
	// Buat data dasar
	baseURL := s.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	maskedToken := utils.MaskToken(s.config.Auth.AccessToken)
	data := entity.NewDashboardData(baseURL, maskedToken)

	// Dapatkan status koneksi WhatsApp
	connectionState, _ := s.whatsApp.GetConnectionStateSafe()
	data.IsConnected = connectionState.IsConnected
	data.ConnectionStatus = string(connectionState.Status)
	data.LastActivity = connectionState.LastActivity
	data.LastActivityFormatted = utils.FormatTimeIndonesia(&connectionState.LastActivity)

	// Jika terhubung, tambahkan informasi koneksi
	if data.IsConnected {
		s.enrichConnectedData(&data, connectionState)
	} else {
		// Jika tidak terhubung, dapatkan informasi QR code
		qrStatus := s.GetQRCodeStatus()
		data.QRCodeAvailable = qrStatus.Available
		data.QRCodeExpired = qrStatus.Expired
		data.QRCodeURL = qrStatus.URL
		data.QRCodeMessage = qrStatus.Message
	}

	return data
}

// enrichConnectedData memperkaya data dashboard dengan informasi koneksi
func (s *DashboardService) enrichConnectedData(data *entity.DashboardData, connectionState client.ConnectionState) {
	// Dapatkan profil info
	deviceInfo := s.whatsApp.GetDeviceInfo()
	connectionInfo := s.whatsApp.GetConnectionInfo()

	// Set waktu koneksi
	s.setConnectionTiming(data, connectionInfo, connectionState)

	// Set informasi profil
	s.setProfileInfo(data, deviceInfo)

	// Dapatkan jumlah pesan terkirim
	messagesSent := s.whatsApp.GetMessagesSent()
	data.MessagesSent = int(messagesSent)

	s.logger.Debug("Retrieved message statistics", utils.Fields{
		"messages_sent": messagesSent,
	})
}

// setConnectionTiming mengatur waktu-waktu terkait koneksi
func (s *DashboardService) setConnectionTiming(data *entity.DashboardData, connectionInfo map[string]interface{}, connectionState client.ConnectionState) {
	// Dapatkan waktu connectedSince yang lebih akurat dari profil WhatsApp
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

		s.logger.Debug("Informasi durasi koneksi dari GetConnectionInfo", utils.Fields{
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

			s.logger.Debug("Fallback ke ConnectionState untuk durasi koneksi", utils.Fields{
				"connected_since": connectionState.ConnectedSince.Format(time.RFC3339),
				"duration":        data.ConnectionDuration,
			})
		} else {
			// Jika masih tidak ada, gunakan LastActivity sebagai pilihan terakhir
			s.logger.Warn("Tidak ada ConnectedSince yang valid, menggunakan LastActivity", utils.Fields{
				"last_activity": connectionState.LastActivity.Format(time.RFC3339),
			})

			data.ConnectedSince = connectionState.LastActivity
			data.ConnectedSinceFormatted = utils.FormatTimeIndonesia(&connectionState.LastActivity)
			data.ConnectionDuration = "Tidak diketahui"
		}
	}
}

// setProfileInfo mengatur informasi profil dalam data dashboard
func (s *DashboardService) setProfileInfo(data *entity.DashboardData, deviceInfo map[string]interface{}) {
	// Set default values untuk mencegah nil
	data.PhoneNumber = "Tidak tersedia"
	data.ContactName = "Tidak tersedia"

	if jid, ok := deviceInfo["id"].(string); ok && jid != "" {
		data.PhoneNumber = utils.FormatWhatsAppNumber(jid)
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
		s.logger.Debug("Menggunakan foto profil", utils.Fields{
			"url": pictureURL,
		})
	} else {
		s.logger.Debug("Foto profil tidak tersedia")
	}
}

// QRStatus adalah struktur data untuk status QR code
type QRStatus struct {
	Available bool
	Expired   bool
	URL       string
	Message   string
}

// GetQRCodeStatus mendapatkan status QR code
func (s *DashboardService) GetQRCodeStatus() QRStatus {
	// Periksa apakah QR handler tersedia
	qrHandler := s.whatsApp.SessionManager.GetQRHandler()
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
		qrURL = fmt.Sprintf("%s/api/qr/image?t=%d", s.config.Server.BaseURL, timestamp.Unix())
	}

	return QRStatus{
		Available: available,
		Expired:   expired,
		URL:       qrURL,
		Message:   message,
	}
}

// RefreshQRCode meminta refresh QR code
func (s *DashboardService) RefreshQRCode() error {
	s.logger.Info("Mencoba mendapatkan QR code baru")

	// Hubungkan ulang WhatsApp untuk mendapatkan QR code baru
	return s.whatsApp.Connect()
}

// DisconnectWhatsApp memutuskan koneksi WhatsApp
func (s *DashboardService) DisconnectWhatsApp() error {
	s.logger.Info("Memutuskan koneksi WhatsApp")

	// Putuskan koneksi WhatsApp
	s.whatsApp.Disconnect()

	// Tunggu sejenak agar status koneksi diperbarui
	time.Sleep(300 * time.Millisecond)

	// Hapus sesi
	return s.whatsApp.SessionManager.ClearSessions()
}
