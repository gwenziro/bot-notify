package storage

import (
	"context"
	"time"

	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// Storage adalah interface untuk penyimpanan data
type Storage interface {
	// Get mengambil nilai dari key yang diberikan
	Get(ctx context.Context, key string) ([]byte, error)

	// Set menyimpan nilai dengan key yang diberikan
	Set(ctx context.Context, key string, value []byte) error

	// SetWithTTL menyimpan nilai dengan key dan TTL yang diberikan
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete menghapus nilai dengan key yang diberikan
	Delete(ctx context.Context, key string) error

	// GetWithPrefix mengembalikan semua key-value dengan prefix yang diberikan
	GetWithPrefix(ctx context.Context, prefix string) (map[string][]byte, error)

	// DeleteWithPrefix menghapus semua key-value dengan prefix yang diberikan
	DeleteWithPrefix(ctx context.Context, prefix string) error

	// Close menutup storage dan membersihkan resource
	Close() error
}

// StorageOptions adalah opsi untuk inisialisasi Storage
type StorageOptions struct {
	Path     string
	InMemory bool
}

// StorageFactory adalah fungsi yang membuat instance Storage baru
type StorageFactory func(options StorageOptions) (Storage, error)

// ErrNotFound adalah error yang dikembalikan ketika key tidak ditemukan
type ErrNotFound struct {
	Key string
}

// Error implementasi interface error untuk ErrNotFound
func (e ErrNotFound) Error() string {
	return "key not found: " + e.Key
}

// IsNotFound memeriksa apakah error adalah ErrNotFound
func IsNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

// Initialize mempersiapkan dan mengembalikan instance Storage yang siap digunakan
func Initialize(cfg *config.Config) (Storage, error) {
	utils.Info("Menginisialisasi storage", utils.Fields{"type": cfg.Storage.Type})

	switch cfg.Storage.Type {
	case "badger":
		return NewBadgerStorage(StorageOptions{
			Path:     cfg.Storage.Path,
			InMemory: cfg.Storage.InMemory,
		})
	default:
		utils.Warn("Tipe storage tidak dikenal, menggunakan badger", utils.Fields{"type": cfg.Storage.Type})
		return NewBadgerStorage(StorageOptions{
			Path:     cfg.Storage.Path,
			InMemory: cfg.Storage.InMemory,
		})
	}
}

var defaultStorage Storage

// GetStorage mengembalikan instance storage yang telah diinisialisasi
func GetStorage() Storage {
	// Jika belum diinisialisasi, kembalikan dummy storage atau nil
	if defaultStorage == nil {
		// Untuk development, kita bisa mengembalikan NoOpStorage
		// atau panic untuk memastikan storage diinisialisasi
		return NewNoOpStorage()
	}
	return defaultStorage
}

// NewNoOpStorage membuat storage yang tidak melakukan apa-apa (untuk pengujian)
func NewNoOpStorage() Storage {
	return &NoOpStorage{}
}

// NoOpStorage adalah implementasi dummy dari Storage interface
type NoOpStorage struct{}

func (s *NoOpStorage) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (s *NoOpStorage) Set(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (s *NoOpStorage) SetWithTTL(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (s *NoOpStorage) Delete(_ context.Context, _ string) error {
	return nil
}

func (s *NoOpStorage) GetWithPrefix(_ context.Context, _ string) (map[string][]byte, error) {
	return nil, nil
}

func (s *NoOpStorage) DeleteWithPrefix(_ context.Context, _ string) error {
	return nil
}

func (s *NoOpStorage) Close() error {
	return nil
}
