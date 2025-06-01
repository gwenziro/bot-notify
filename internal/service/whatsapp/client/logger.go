package client

import (
	"strings"

	"github.com/gwenziro/bot-notify/internal/utils"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WhatsmeowLogger adalah implementasi yang sesuai dengan interface waLog.Logger
type WhatsmeowLogger struct {
	logger utils.LogrusEntry
	name   string
	// List of patterns to filter from logs
	filterPatterns []string
}

// NewWhatsmeowLogger membuat instance baru WhatsmeowLogger
func NewWhatsmeowLogger(logger utils.LogrusEntry) *WhatsmeowLogger {
	// Common error patterns to filter out
	filterPatterns := []string{
		"status@broadcast",
		"database is locked",
		"SQLITE_BUSY",
		"Failed to migrate signal store",
		"Failed to save push name",
		"Error decrypting message",
		"Failed to sync app state",
		"Failed to get prekey for retry receipt",
	}

	return &WhatsmeowLogger{
		logger:         logger,
		name:           "whatsmeow",
		filterPatterns: filterPatterns,
	}
}

// shouldFilter memeriksa apakah pesan log perlu difilter
func (l *WhatsmeowLogger) shouldFilter(message string) bool {
	for _, pattern := range l.filterPatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// formatMessage memformat pesan log dengan argumen
func (l *WhatsmeowLogger) formatMessage(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return utils.Sprintf(format, args...)
}

// Debugf implementasi waLog.Logger
func (l *WhatsmeowLogger) Debugf(format string, args ...interface{}) {
	message := l.formatMessage(format, args...)
	if l.shouldFilter(message) {
		return
	}
	l.logger.Debugf(format, args...)
}

// Infof implementasi waLog.Logger
func (l *WhatsmeowLogger) Infof(format string, args ...interface{}) {
	message := l.formatMessage(format, args...)
	if l.shouldFilter(message) {
		return
	}
	l.logger.Infof(format, args...)
}

// Warnf implementasi waLog.Logger
func (l *WhatsmeowLogger) Warnf(format string, args ...interface{}) {
	message := l.formatMessage(format, args...)
	if l.shouldFilter(message) {
		return
	}
	l.logger.Warnf(format, args...)
}

// Errorf implementasi waLog.Logger
func (l *WhatsmeowLogger) Errorf(format string, args ...interface{}) {
	message := l.formatMessage(format, args...)
	if l.shouldFilter(message) {
		return
	}
	l.logger.Errorf(format, args...)
}

// Sub implementasi waLog.Logger
func (l *WhatsmeowLogger) Sub(module string) waLog.Logger {
	return &WhatsmeowLogger{
		logger:         l.logger.WithField("sub", module),
		name:           module,
		filterPatterns: l.filterPatterns,
	}
}
