package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// BadgerStorage adalah implementasi Storage menggunakan BadgerDB
type BadgerStorage struct {
	db     *badger.DB
	logger utils.LogrusEntry
}

// NewBadgerStorage membuat instance baru BadgerStorage
func NewBadgerStorage(options StorageOptions) (Storage, error) {
	// Gunakan nama modul yang konsisten untuk BadgerDB
	log := utils.ForModule("badger")

	var db *badger.DB
	var err error

	// Konfigurasi optimasi untuk shared hosting dengan performa maksimal
	opts := badger.DefaultOptions("")

	if options.InMemory {
		// Mode in-memory - tetap gunakan memori maksimal untuk performa
		opts = opts.WithInMemory(true)
		log.Info("Menggunakan BadgerDB dalam mode in-memory")
	} else {
		// Mode file
		if err := os.MkdirAll(options.Path, 0755); err != nil {
			return nil, fmt.Errorf("gagal membuat direktori penyimpanan: %w", err)
		}

		dataPath := filepath.Join(options.Path, "data")
		opts = opts.WithDir(dataPath).WithValueDir(dataPath)

		// Coba hapus lock file jika ada
		lockFile := filepath.Join(dataPath, "LOCK")
		if _, err := os.Stat(lockFile); err == nil {
			log.Warn("Ditemukan file LOCK eksisting, mencoba menghapus...")
			if err := os.Remove(lockFile); err != nil {
				log.WithError(err).Warn("Gagal menghapus file LOCK, mungkin masih digunakan")
			}
		}

		log.WithField("path", dataPath).Info("Menggunakan BadgerDB dalam mode file")
	}

	// Optimasi performa tanpa terlalu ketat dengan resource

	// 1. Pengaturan memori optimal untuk performa baik
	opts = opts.WithMemTableSize(32 << 20)      // 32MB (default 64MB) - cukup besar untuk performa, tapi tidak berlebihan
	opts = opts.WithValueLogFileSize(256 << 20) // 256MB (default 1GB) - lebih besar dari sebelumnya
	opts = opts.WithNumMemtables(2)             // Lebih dari 1 untuk performa yang lebih baik
	opts = opts.WithNumLevelZeroTables(3)       // Sedikit lebih besar dari sebelumnya
	opts = opts.WithNumLevelZeroTablesStall(5)  // Lebih besar untuk mengurangi stall

	// 2. Pengaturan cache untuk performa baca yang lebih baik
	opts = opts.WithBlockCacheSize(32 << 20) // 32MB untuk block cache - meningkatkan performa pembacaan
	opts = opts.WithIndexCacheSize(16 << 20) // 16MB untuk index cache - meningkatkan lookup

	// 3. Optimasi operasi disk yang sesuai untuk shared hosting
	opts = opts.WithSyncWrites(false)      // Performa lebih penting daripada durabilitas 100%
	opts = opts.WithCompression(1)         // 1 = Snappy compression (hemat ruang disk dengan overhead rendah)
	opts = opts.WithDetectConflicts(false) // Matikan deteksi konflik untuk performa lebih baik

	// 4. Pengaturan logger
	opts = opts.WithLogger(utils.Log)
	opts = opts.WithLoggingLevel(badger.WARNING) // Hanya log warning/error

	// Atur batas waktu pembukaan database
	timeout := 5 * time.Second
	openCh := make(chan error, 1)

	// Buka database dengan timeout
	go func() {
		db, err = badger.Open(opts)
		openCh <- err
	}()

	// Tunggu database terbuka atau timeout
	select {
	case err := <-openCh:
		if err != nil {
			return nil, fmt.Errorf("gagal membuka BadgerDB: %w", err)
		}
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout saat membuka BadgerDB setelah %v", timeout)
	}

	// Jalankan garbage collection secara terpisah
	go runGarbageCollection(db, log)

	return &BadgerStorage{
		db:     db,
		logger: log,
	}, nil
}

// runGarbageCollection menjalankan GC pada interval tertentu
func runGarbageCollection(db *badger.DB, log utils.LogrusEntry) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Lakukan value log GC setelah interval waktu
		err := db.RunValueLogGC(0.5) // 50% threshold - lebih agresif
		if err != nil {
			if err != badger.ErrNoRewrite {
				log.WithError(err).Warn("Garbage collection error")
			}
			continue
		}

		// Jika GC berhasil, coba lagi setelah jeda singkat
		time.Sleep(1 * time.Second)
		_ = db.RunValueLogGC(0.5)
	}
}

// Get mengambil nilai dari key yang diberikan
func (s *BadgerStorage) Get(ctx context.Context, key string) ([]byte, error) {
	var valCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return ErrNotFound{Key: key}
		} else if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return valCopy, nil
}

// Set menyimpan nilai dengan key yang diberikan
func (s *BadgerStorage) Set(ctx context.Context, key string, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value)
		return txn.SetEntry(entry)
	})
}

// SetWithTTL menyimpan nilai dengan key dan TTL yang diberikan
func (s *BadgerStorage) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// Delete menghapus nilai dengan key yang diberikan
func (s *BadgerStorage) Delete(ctx context.Context, key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// GetWithPrefix mengembalikan semua key-value dengan prefix yang diberikan
func (s *BadgerStorage) GetWithPrefix(ctx context.Context, prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key())

			var value []byte
			err := item.Value(func(val []byte) error {
				value = append([]byte{}, val...)
				return nil
			})
			if err != nil {
				return err
			}

			result[key] = value
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteWithPrefix menghapus semua key-value dengan prefix yang diberikan
func (s *BadgerStorage) DeleteWithPrefix(ctx context.Context, prefix string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefixBytes := []byte(prefix)
		var keysToDelete [][]byte

		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			keysToDelete = append(keysToDelete, key)
		}

		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close menutup storage dan membersihkan resource
func (s *BadgerStorage) Close() error {
	s.logger.Info("Menutup koneksi BadgerDB")
	return s.db.Close()
}
