package log

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/repository"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// LogService menyediakan operasi terkait log
type LogService struct {
	repository *repository.LogRepository
	logger     utils.LogrusEntry
}

// NewLogService membuat service log baru
func NewLogService(repository *repository.LogRepository, logger utils.LogrusEntry) *LogService {
	return &LogService{
		repository: repository,
		logger:     logger.WithField("component", "log-service"),
	}
}

// CreateLog membuat dan menyimpan log baru
func (s *LogService) CreateLog(level, source, message string, data map[string]interface{}) error {
	log := model.NewLog(level, source, message, data)
	return s.repository.SaveLog(log)
}

// GetPaginatedLogs mendapatkan logs dengan paginasi dan filter
func (s *LogService) GetPaginatedLogs(page, limit int, level, source, search, from, to string) ([]*model.Log, int, int, error) {
	// Validasi parameter
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25 // Default limit
	}

	// Ambil logs dari repository dengan filter
	logs, totalLogs, err := s.repository.GetLogsByFilter(page, limit, level, source, search, from, to)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("gagal mendapatkan logs: %w", err)
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalLogs) / float64(limit)))

	return logs, totalLogs, totalPages, nil
}

// ClearAllLogs menghapus seluruh data log
func (s *LogService) ClearAllLogs() error {
	err := s.repository.ClearAllLogs()
	if err != nil {
		return fmt.Errorf("gagal menghapus logs: %w", err)
	}

	// Tambahkan log sistem baru bahwa logs telah dibersihkan
	s.CreateLog(
		string(model.LogLevelInfo),
		string(model.LogSourceSystem),
		"Semua log telah dibersihkan",
		map[string]interface{}{
			"action": "clear_logs",
		},
	)

	return nil
}

// ExportLogs mengekspor logs dalam format tertentu
func (s *LogService) ExportLogs(format, level, source, search, from, to string) ([]byte, string, string, error) {
	// Ambil semua log dengan filter tanpa paginasi (limit besar)
	logs, _, _, err := s.GetPaginatedLogs(1, 10000, level, source, search, from, to)
	if err != nil {
		return nil, "", "", fmt.Errorf("gagal mendapatkan logs untuk ekspor: %w", err)
	}

	// Tentukan nama file berdasarkan timestamp
	timestamp := time.Now().Format("20060102_150405")
	var fileName, contentType string
	var fileBytes []byte

	// Ekspor berdasarkan format
	switch format {
	case "csv":
		fileBytes, err = s.exportToCSV(logs)
		fileName = fmt.Sprintf("logs_export_%s.csv", timestamp)
		contentType = "text/csv"
	case "json":
		fileBytes, err = s.exportToJSON(logs)
		fileName = fmt.Sprintf("logs_export_%s.json", timestamp)
		contentType = "application/json"
	default:
		return nil, "", "", fmt.Errorf("format ekspor tidak didukung: %s", format)
	}

	if err != nil {
		return nil, "", "", fmt.Errorf("gagal mengekspor logs: %w", err)
	}

	// Catat aktivitas ekspor log ke log
	s.CreateLog(
		string(model.LogLevelInfo),
		string(model.LogSourceSystem),
		fmt.Sprintf("Log diekspor dalam format %s", format),
		map[string]interface{}{
			"action":        "export_logs",
			"format":        format,
			"filter_level":  level,
			"filter_source": source,
			"count":         len(logs),
		},
	)

	return fileBytes, fileName, contentType, nil
}

// exportToCSV mengekspor log dalam format CSV
func (s *LogService) exportToCSV(logs []*model.Log) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// Tulis header CSV
	headers := []string{"Timestamp", "Level", "Source", "Message", "Data"}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("gagal menulis header CSV: %w", err)
	}

	// Tulis data log
	for _, log := range logs {
		var dataStr string
		if log.Data != nil {
			dataBytes, err := json.Marshal(log.Data)
			if err == nil {
				dataStr = string(dataBytes)
			}
		}

		row := []string{
			log.Timestamp.Format(time.RFC3339),
			log.Level,
			log.Source,
			log.Message,
			dataStr,
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("gagal menulis baris CSV: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("kesalahan saat menulis CSV: %w", err)
	}

	return buf.Bytes(), nil
}

// exportToJSON mengekspor log dalam format JSON
func (s *LogService) exportToJSON(logs []*model.Log) ([]byte, error) {
	return json.MarshalIndent(logs, "", "  ")
}
