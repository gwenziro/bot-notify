package model

import (
	"time"
)

// Log mewakili sebuah entri log dalam sistem
type Log struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// LogLevel mendefinisikan level log yang tersedia
type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
)

// LogSource mendefinisikan sumber log
type LogSource string

const (
	LogSourceSystem   LogSource = "SYSTEM"
	LogSourceWhatsApp LogSource = "WHATSAPP"
	LogSourceAPI      LogSource = "API"
	LogSourceWeb      LogSource = "WEB"
	LogSourceDatabase LogSource = "DATABASE"
)

// LogListResponse adalah respons untuk request list logs
type LogListResponse struct {
	Logs       []*Log `json:"logs"`
	Page       int    `json:"page"`
	TotalPages int    `json:"totalPages"`
	TotalLogs  int    `json:"totalLogs"`
	Success    bool   `json:"success"`
}

// LogRequest adalah request untuk operasi log
type LogRequest struct {
	Level   string                 `json:"level"`
	Source  string                 `json:"source"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NewLog membuat log baru dengan timestamp saat ini
func NewLog(level, source, message string, data map[string]interface{}) *Log {
	return &Log{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
		Data:      data,
	}
}
