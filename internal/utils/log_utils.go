package utils

import (
	"sort"

	"github.com/gwenziro/bot-notify/internal/api/model"
)

// SortLogs mengurutkan log berdasarkan timestamp (terbaru dulu)
func SortLogs(logs []*model.Log) {
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})
}

// GetLogLevels mengembalikan semua level log yang tersedia
func GetLogLevels() []string {
	return []string{
		string(model.LogLevelDebug),
		string(model.LogLevelInfo),
		string(model.LogLevelWarning),
		string(model.LogLevelError),
		string(model.LogLevelFatal),
	}
}

// GetLogSources mengembalikan semua sumber log yang tersedia
func GetLogSources() []string {
	return []string{
		string(model.LogSourceSystem),
		string(model.LogSourceWhatsApp),
		string(model.LogSourceAPI),
		string(model.LogSourceWeb),
		string(model.LogSourceDatabase),
	}
}

// GetLogLevelClass mengembalikan kelas CSS untuk level log tertentu
func GetLogLevelClass(level string) string {
	switch level {
	case string(model.LogLevelDebug):
		return "info"
	case string(model.LogLevelInfo):
		return "success"
	case string(model.LogLevelWarning):
		return "warning"
	case string(model.LogLevelError), string(model.LogLevelFatal):
		return "danger"
	default:
		return "secondary"
	}
}
