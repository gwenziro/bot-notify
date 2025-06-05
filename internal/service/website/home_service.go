package website

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web/entity"
)

// HomeService menyediakan fungsionalitas untuk halaman beranda
type HomeService struct {
	config *config.Config
	logger utils.LogrusEntry
}

// NewHomeService membuat instance service beranda baru
func NewHomeService(cfg *config.Config, logger utils.LogrusEntry) *HomeService {
	return &HomeService{
		config: cfg,
		logger: logger.WithField("component", "home-service"),
	}
}

// GetHomePageData menyiapkan data untuk halaman beranda
func (s *HomeService) GetHomePageData() entity.HomePageData {
	// Buat data untuk halaman home
	data := entity.HomePageData{
		Title:       "WhatsApp Bot Notify",
		Description: "Bot WhatsApp Kirim Pesan Realtime",
		Version:     "1.0.0",
		CurrentYear: time.Now().Year(),
	}

	// Dapatkan base URL dari konfigurasi
	baseURL := s.config.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Tetapkan API URL dengan base URL yang sesuai
	data.ApiBaseURL = baseURL

	// Mungkin di produksi kita ingin override dengan domain khusus API
	if s.config.Server.Environment == "production" {
		if s.config.Server.ApiDomain != "" {
			data.ApiBaseURL = s.config.Server.ApiDomain
		}
	}

	s.logger.Debug("Home page data prepared", utils.Fields{
		"api_base_url": data.ApiBaseURL,
	})

	return data
}
