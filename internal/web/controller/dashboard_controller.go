package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// DashboardController menangani halaman dashboard
type DashboardController struct {
	config       *config.Config
	whatsApp     *client.Client
	logger       utils.LogrusEntry
	startTime    time.Time
	messagesSent int

	// Cache untuk mengurangi panggilan ke GetGroups()
	groupsCache      []*types.GroupInfo
	groupsCacheTime  time.Time
	groupsCacheMutex sync.RWMutex
	groupsCacheTTL   time.Duration
}

// NewDashboardController membuat instance baru DashboardController
func NewDashboardController(cfg *config.Config, whatsClient *client.Client, logger utils.LogrusEntry) *DashboardController {
	return &DashboardController{
		config:         cfg,
		whatsApp:       whatsClient,
		logger:         logger.WithField("component", "dashboard-controller"),
		startTime:      time.Now(),
		messagesSent:   0,
		groupsCacheTTL: 5 * time.Minute, // Cache grup selama 5 menit
	}
}

// IncrementMessageCounter menambah counter pesan terkirim
func (c *DashboardController) IncrementMessageCounter() {
	c.messagesSent++
}

// GetMessageCount mengembalikan jumlah pesan terkirim
func (c *DashboardController) GetMessageCount() int {
	return c.messagesSent
}

// formatUptime menghasilkan string uptime yang mudah dibaca
func formatUptime(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}

	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

// getGroupsWithCache mendapatkan daftar grup dengan cache untuk menghindari panggilan berulang
func (c *DashboardController) getGroupsWithCache() (int, error) {
	// Gunakan read lock untuk memeriksa cache
	c.groupsCacheMutex.RLock()
	cacheValid := !c.groupsCacheTime.IsZero() && time.Since(c.groupsCacheTime) < c.groupsCacheTTL
	groupCount := len(c.groupsCache)
	c.groupsCacheMutex.RUnlock()

	// Jika cache masih valid, gunakan saja
	if cacheValid {
		c.logger.Debug("Menggunakan cache grup yang sudah ada")
		return groupCount, nil
	}

	// Cache tidak valid, perlu mengambil data baru dengan write lock
	c.groupsCacheMutex.Lock()
	defer c.groupsCacheMutex.Unlock()

	// Periksa lagi apakah cache sudah diperbarui oleh goroutine lain
	if !c.groupsCacheTime.IsZero() && time.Since(c.groupsCacheTime) < c.groupsCacheTTL {
		return len(c.groupsCache), nil
	}

	// Ambil data grup baru
	c.logger.Info("Memperbarui cache grup")
	groups, err := c.whatsApp.GetGroups()
	if err != nil {
		return 0, err
	}

	// Perbarui cache
	c.groupsCache = groups
	c.groupsCacheTime = time.Now()

	return len(groups), nil
}

// DashboardPage menampilkan halaman dashboard utama
func (c *DashboardController) DashboardPage(ctx *fiber.Ctx) error {
	c.logger.Debug("Rendering dashboard page")

	// Dapatkan status koneksi WhatsApp
	connectionState := c.whatsApp.GetConnectionState()

	// Hitung uptime aplikasi
	uptime := formatUptime(time.Since(c.startTime))

	// Dapatkan data statistik tambahan jika terhubung
	var groupsCount int
	messagesSent := c.messagesSent

	if connectionState.IsConnected {
		// Gunakan cache untuk menghindari pengambilan grup berulang
		var err error
		groupsCount, err = c.getGroupsWithCache()
		if err != nil {
			c.logger.Warn("Gagal mendapatkan jumlah grup", utils.Fields{
				"error": err.Error(),
			})
			// Tetap lanjutkan meskipun error, hanya jumlah grup akan 0
		}
	}

	// Persiapkan data untuk template
	baseURL := c.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Mask token untuk contoh API
	maskedToken := maskToken(c.config.Auth.AccessToken)

	// Buat data untuk template
	data := fiber.Map{
		"Title":             "Dashboard",
		"CurrentYear":       time.Now().Year(),
		"IsConnected":       connectionState.IsConnected,
		"ConnectionStatus":  string(connectionState.Status),
		"ConnectedSince":    connectionState.LastActivity,
		"ConnectionRetries": connectionState.ConnectionRetries,
		"DeviceName":        "WhatsApp Web", // Default, bisa diubah jika ada info device
		"MessagesSent":      messagesSent,
		"GroupsCount":       groupsCount,
		"Uptime":            uptime,
		"BaseURL":           baseURL,
		"MaskedToken":       maskedToken,
		"ConnectionState":   connectionState,
		"ActivePage":        "dashboard",                 // Untuk highlight menu aktif di sidebar
		"WhatsAppConnected": connectionState.IsConnected, // Untuk status di sidebar
	}

	// Tambahkan log untuk debugging dengan level yang lebih rendah
	c.logger.Debug("Dashboard data prepared", utils.Fields{
		"is_connected":     connectionState.IsConnected,
		"connection_state": connectionState.Status,
		"groups_count":     groupsCount,
		"messages_sent":    messagesSent,
	})

	// Render dashboard dengan data
	return ctx.Render("dashboard", data)
}

// maskToken menyembunyikan sebagian token untuk keamanan
func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}

	// Tampilkan hanya 4 karakter pertama dan 4 terakhir
	return token[:4] + "..." + token[len(token)-4:]
}
