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

// Fungsi NewStorage untuk membuat instance storage baru
// Gunakan factory pattern untuk mendukung berbagai implementasi storage
func NewStorage(cfg *config.Config) (Storage, error) {
	// Gunakan tipe storage dari konfigurasi
	switch cfg.Storage.Type {
	case "badger":
		return NewBadgerStorage(StorageOptions{
			Path:     cfg.Storage.Path,
			InMemory: cfg.Storage.InMemory,
		})
	case "memory":
		storage := NewMemoryStorage()
		return storage, nil
	default:
		// Default ke badger storage
		return NewBadgerStorage(StorageOptions{
			Path:     cfg.Storage.Path,
			InMemory: cfg.Storage.InMemory,
		})
	}
}

// MemoryStorage implements Storage using an in-memory map
type MemoryStorage struct {
	data map[string][]byte
	ttl  map[string]time.Time
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
		ttl:  make(map[string]time.Time),
	}
}

// Get retrieves a value by key from memory
func (s *MemoryStorage) Get(ctx context.Context, key string) ([]byte, error) {
	if value, ok := s.data[key]; ok {
		// Check if the key has expired
		if expiry, hasExpiry := s.ttl[key]; hasExpiry && time.Now().After(expiry) {
			delete(s.data, key)
			delete(s.ttl, key)
			return nil, ErrNotFound{Key: key}
		}
		return value, nil
	}
	return nil, ErrNotFound{Key: key}
}

// Set stores a value with the given key in memory
func (s *MemoryStorage) Set(ctx context.Context, key string, value []byte) error {
	s.data[key] = value
	delete(s.ttl, key) // Remove any TTL for this key
	return nil
}

// SetWithTTL stores a value with the given key and TTL in memory
func (s *MemoryStorage) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	s.data[key] = value
	s.ttl[key] = time.Now().Add(ttl)
	return nil
}

// Delete removes a key-value pair from memory
func (s *MemoryStorage) Delete(ctx context.Context, key string) error {
	delete(s.data, key)
	delete(s.ttl, key)
	return nil
}

// GetWithPrefix returns all key-value pairs with the given prefix
func (s *MemoryStorage) GetWithPrefix(ctx context.Context, prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	now := time.Now()

	for k, v := range s.data {
		// Check if key has the prefix and is not expired
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			if expiry, hasExpiry := s.ttl[k]; !hasExpiry || now.Before(expiry) {
				result[k] = v
			}
		}
	}

	return result, nil
}

// DeleteWithPrefix removes all key-value pairs with the given prefix
func (s *MemoryStorage) DeleteWithPrefix(ctx context.Context, prefix string) error {
	for k := range s.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(s.data, k)
			delete(s.ttl, k)
		}
	}
	return nil
}

// Close is a no-op for memory storage
func (s *MemoryStorage) Close() error {
	// No resources to clean up for in-memory storage
	return nil
}
