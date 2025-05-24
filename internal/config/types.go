package config

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
)

// Config adalah struktur konfigurasi utama
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	WhatsApp WhatsAppConfig `yaml:"whatsapp"`
	Auth     AuthConfig     `yaml:"auth"`
	Storage  StorageConfig  `yaml:"storage"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig berisi konfigurasi untuk web server
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	BaseURL         string        `yaml:"base_url"`
}

// WhatsAppConfig berisi konfigurasi untuk layanan WhatsApp
type WhatsAppConfig struct {
	StoreDir    string        `yaml:"store_dir"`
	QrCodeDir   string        `yaml:"qr_code_dir"`
	MaxRetry    int           `yaml:"max_retry"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// AuthConfig berisi konfigurasi untuk autentikasi
type AuthConfig struct {
	TokenSecret  string        `yaml:"token_secret"`
	AccessToken  string        `yaml:"access_token"`
	TokenExpiry  time.Duration `yaml:"token_expiry"`
	SessionDir   string        `yaml:"session_dir"`
	CookieName   string        `yaml:"cookie_name"`
	CookieMaxAge int           `yaml:"cookie_max_age"`
}

// StorageConfig berisi konfigurasi untuk penyimpanan data
type StorageConfig struct {
	Type     string `yaml:"type"`
	Path     string `yaml:"path"`
	InMemory bool   `yaml:"in_memory"`
}

// LoggingConfig adalah konfigurasi untuk logger
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// GetLogConfig mengkonversi LoggingConfig ke utils.LogConfig
func (cfg *Config) GetLogConfig() *utils.LogConfig {
	return &utils.LogConfig{
		Level:      cfg.Logging.Level,
		File:       cfg.Logging.File,
		MaxSize:    cfg.Logging.MaxSize,
		MaxAge:     cfg.Logging.MaxAge,
		MaxBackups: cfg.Logging.MaxBackups,
		Compress:   cfg.Logging.Compress,
	}
}
