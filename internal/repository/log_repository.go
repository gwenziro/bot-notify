package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// LogRepository menangani operasi penyimpanan log
type LogRepository struct {
	storage storage.Storage
	helper  *storage.Helper
	logger  utils.LogrusEntry
}

// NewLogRepository membuat repository log baru
func NewLogRepository(store storage.Storage, logger utils.LogrusEntry) *LogRepository {
	helper := storage.NewHelper(store, "logs")
	return &LogRepository{
		storage: store,
		helper:  helper,
		logger:  logger.WithField("component", "log-repository"),
	}
}

// SaveLog menyimpan log ke penyimpanan
func (r *LogRepository) SaveLog(log *model.Log) error {
	ctx := context.Background()

	// Generate ID jika belum ada
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	// Simpan log dengan key berdasarkan waktu untuk urutan kronologis
	// Format: logs:{timestamp}:{level}:{id}
	timeKey := log.Timestamp.Format("20060102150405.000000")
	key := fmt.Sprintf("%s:%s:%s", timeKey, strings.ToLower(log.Level), log.ID)

	// Simpan menggunakan helper JSON
	err := r.helper.SetJSON(ctx, key, log)
	if err != nil {
		return fmt.Errorf("gagal menyimpan log: %w", err)
	}

	return nil
}

// GetLogsByFilter mendapatkan log berdasarkan filter
func (r *LogRepository) GetLogsByFilter(page, limit int, level, source, search, from, to string) ([]*model.Log, int, error) {
	ctx := context.Background()

	// Get all logs with prefix
	data, err := r.helper.GetAllWithPrefix(ctx, "")
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mendapatkan logs: %w", err)
	}

	// Parse semua log dan terapkan filter
	var allLogs []*model.Log
	for _, v := range data {
		var log model.Log
		if err := json.Unmarshal(v, &log); err != nil {
			r.logger.WithError(err).Warn("Gagal parse log entry")
			continue
		}

		// Filter berdasarkan level
		if level != "" && !strings.EqualFold(log.Level, level) {
			continue
		}

		// Filter berdasarkan source
		if source != "" && !strings.EqualFold(log.Source, source) {
			continue
		}

		// Filter berdasarkan teks yang dicari di pesan
		if search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(search)) {
			continue
		}

		// Filter berdasarkan tanggal from
		if from != "" {
			fromDate, err := time.Parse("2006-01-02", from)
			if err == nil && log.Timestamp.Before(fromDate) {
				continue
			}
		}

		// Filter berdasarkan tanggal to
		if to != "" {
			toDate, err := time.Parse("2006-01-02", to)
			if err == nil {
				// Tambahkan 1 hari ke to date untuk mencakup seluruh hari
				toDate = toDate.Add(24 * time.Hour)
				if log.Timestamp.After(toDate) {
					continue
				}
			}
		}

		// Log sesuai filter, tambahkan ke hasil
		allLogs = append(allLogs, &log)
	}

	// Hitung total logs yang sesuai filter
	totalLogs := len(allLogs)

	// Terapkan paginasi
	startIdx := (page - 1) * limit
	if startIdx >= totalLogs {
		return []*model.Log{}, totalLogs, nil
	}

	endIdx := startIdx + limit
	if endIdx > totalLogs {
		endIdx = totalLogs
	}

	// Sort logs by timestamp (descending - newest first)
	utils.SortLogs(allLogs)

	return allLogs[startIdx:endIdx], totalLogs, nil
}

// ClearAllLogs menghapus seluruh data log
func (r *LogRepository) ClearAllLogs() error {
	ctx := context.Background()
	return r.helper.DeleteAllWithPrefix(ctx, "")
}

// GetTotalLogCount mengembalikan jumlah total log dalam sistem
func (r *LogRepository) GetTotalLogCount() (int, error) {
	ctx := context.Background()

	data, err := r.helper.GetAllWithPrefix(ctx, "")
	if err != nil {
		return 0, fmt.Errorf("gagal mendapatkan jumlah log: %w", err)
	}

	return len(data), nil
}
