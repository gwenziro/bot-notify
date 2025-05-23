package session

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Manager menangani sesi WhatsApp dan QR code
type Manager struct {
	config      *config.Config
	logger      utils.LogrusEntry
	deviceStore *sqlstore.Container
	qrHandler   *QRHandler
	client      interface{} // Akan disimpan referensi ke client
}

// NewManager membuat instance baru Manager
func NewManager(cfg *config.Config, logger utils.LogrusEntry, deviceStore *sqlstore.Container) *Manager {
	manager := &Manager{
		config:      cfg,
		logger:      logger.WithField("component", "session-manager"),
		deviceStore: deviceStore,
	}

	// Buat QR handler
	qrCodePath := filepath.Join(cfg.WhatsApp.QrCodeDir, "latest-qr.png")
	manager.qrHandler = NewQRHandler(cfg, logger, qrCodePath)

	return manager
}

// SetClient menyimpan referensi ke client
func (m *Manager) SetClient(client interface{}) {
	m.client = client
}

// SetupQRCodeListener mengatur callback untuk menangani QR code
func (m *Manager) SetupQRCodeListener() {
	// Dapatkan client yang telah dicast ke interface yang sesuai
	client, ok := m.client.(interface {
		RegisterCallback(string, func(interface{}))
	})

	if !ok {
		m.logger.Error("Client tidak mendukung RegisterCallback")
		return
	}

	// Mendaftarkan callback untuk QR code yang akan memanggil QR handler
	client.RegisterCallback("QRCode", func(data interface{}) {
		qrCode, ok := data.(string)
		if !ok || qrCode == "" {
			m.logger.Warn("QR code tidak valid")
			return
		}

		err := m.qrHandler.ProcessQRCode(qrCode)
		if err != nil {
			m.logger.WithError(err).Error("Gagal memproses QR code")
		}
	})

	m.logger.Info("QR code listener berhasil diatur")
}

// GetDevice mengembalikan device store untuk koneksi WhatsApp
func (m *Manager) GetDevice() (*store.Device, error) {
	ctx := context.Background()
	devices, err := m.deviceStore.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan devices: %w", err)
	}

	if len(devices) == 0 {
		m.logger.Info("Tidak ada device yang disimpan, membuat device baru")
		return m.deviceStore.NewDevice(), nil
	}

	// Gunakan device pertama yang ditemukan
	m.logger.WithFields(utils.Fields{
		"id":   devices[0].ID.String(),
		"name": devices[0].PushName,
	}).Info("Menggunakan device yang ada")
	return devices[0], nil
}

// ClearSessions menghapus semua sesi yang disimpan
func (m *Manager) ClearSessions() error {
	m.logger.Warn("Menghapus semua sesi WhatsApp")

	ctx := context.Background()

	// Tutup device store saat ini
	if m.deviceStore != nil {
		m.deviceStore.Close()
	}

	// Buat device store baru dengan database yang sama
	dbPath := fmt.Sprintf("%s/store.db", m.config.WhatsApp.StoreDir)
	// Gunakan driver sqlite dari modernc.org
	deviceStore, err := sqlstore.New(ctx, "sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath), waLog.Noop)
	if err != nil {
		return fmt.Errorf("gagal membuat device store baru: %w", err)
	}

	// Hapus semua device
	devices, err := deviceStore.GetAllDevices(ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan devices: %w", err)
	}

	for _, device := range devices {
		err := deviceStore.DeleteDevice(ctx, device)
		if err != nil {
			m.logger.WithFields(utils.Fields{
				"id":    device.ID.String(),
				"error": err,
			}).Error("Gagal menghapus device")
		} else {
			m.logger.WithField("id", device.ID.String()).Info("Berhasil menghapus device")
		}
	}

	// Perbarui device store di SessionManager
	m.deviceStore = deviceStore

	// Hapus QR code
	m.qrHandler.ClearQRCode()

	return nil
}

// GetQRHandler mengembalikan QR handler untuk digunakan eksternal
func (m *Manager) GetQRHandler() *QRHandler {
	return m.qrHandler
}
