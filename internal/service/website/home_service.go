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
	// Gunakan CleanBaseURL untuk memastikan tidak ada port yang terlihat
	baseURL := utils.CleanBaseURL(s.config.Server.BaseURL)

	return entity.HomePageData{
		Title:       "Bot Notify - Kirim Notifikasi WhatsApp dengan Mudah",
		Description: "Bot Notify memungkinkan integrasi notifikasi WhatsApp dengan aplikasi Anda dengan mudah melalui API.",
		CurrentYear: time.Now().Year(),
		ApiBaseURL:  baseURL,
	}
}
