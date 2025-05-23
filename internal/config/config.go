package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
	"gopkg.in/yaml.v2"
)

// LoadDefault memuat konfigurasi dari path default
func LoadDefault() (*Config, error) {
	// Gunakan ProjectRoot langsung untuk mengakses direktori config
	configDir := filepath.Join(utils.ProjectRoot, "config")
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Jika file tidak ada, buat dengan config default
			cfg := defaultConfig()
			yamlData, err := yaml.Marshal(cfg)
			if err != nil {
				return nil, fmt.Errorf("gagal marshal config default: %w", err)
			}

			// Pastikan direktori config ada
			if err := utils.EnsureDirectoryExists(configDir); err != nil {
				return nil, fmt.Errorf("gagal membuat direktori config: %w", err)
			}

			// Simpan ke path konfigurasi
			err = ioutil.WriteFile(configPath, yamlData, 0644)
			if err != nil {
				return nil, fmt.Errorf("gagal menulis config default ke file: %w", err)
			}
			fmt.Printf("File konfigurasi baru dibuat di '%s'\n", configPath)
			return cfg, nil
		}
		return nil, fmt.Errorf("gagal membaca file konfigurasi: %w", err)
	}

	// Parse YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("gagal parse YAML: %w", err)
	}

	return &config, nil
}

// LoadFromPath memuat konfigurasi dari path yang ditentukan
func LoadFromPath(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file konfigurasi: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("gagal parse YAML: %w", err)
	}

	return &config, nil
}

// defaultConfig mengembalikan konfigurasi default
func defaultConfig() *Config {
	// Akses direktori langsung menggunakan ProjectRoot
	dataDir := filepath.Join(utils.ProjectRoot, "data")
	logsDir := filepath.Join(utils.ProjectRoot, "logs")

	return &Config{
		Server: ServerConfig{
			Host:            "127.0.0.1",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			BaseURL:         "http://localhost:8080",
		},
		WhatsApp: WhatsAppConfig{
			StoreDir:    filepath.Join(dataDir, "whatsapp"),
			QrCodeDir:   filepath.Join(dataDir, "qrcodes"),
			MaxRetry:    5,
			RetryDelay:  5 * time.Second,
			IdleTimeout: 30 * time.Minute,
		},
		Auth: AuthConfig{
			TokenSecret:  "change-this-to-secure-random-string",
			AccessToken:  "change-this-to-your-api-token",
			TokenExpiry:  24 * time.Hour,
			SessionDir:   filepath.Join(dataDir, "sessions"),
			CookieName:   "whatsmeow_session",
			CookieMaxAge: 86400,
		},
		Storage: StorageConfig{
			Type:     "badger",
			Path:     filepath.Join(dataDir, "storage"),
			InMemory: false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       filepath.Join(logsDir, "app.log"),
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
	}
}

// SaveToFile menyimpan konfigurasi ke file
func SaveToFile(cfg *Config, path string) error {
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("gagal marshal config: %w", err)
	}

	// Pastikan direktori ada
	dir := filepath.Dir(path)
	if err := utils.EnsureDirectoryExists(dir); err != nil {
		return fmt.Errorf("gagal membuat direktori: %w", err)
	}

	// Simpan ke file
	err = ioutil.WriteFile(path, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis config ke file: %w", err)
	}

	return nil
}
