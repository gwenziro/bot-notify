package client

import (
	"github.com/gwenziro/bot-notify/internal/utils"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WhatsmeowLogger adalah implementasi yang sesuai dengan interface waLog.Logger
type WhatsmeowLogger struct {
	logger utils.LogrusEntry
	name   string
}

// NewWhatsmeowLogger membuat instance baru WhatsmeowLogger
func NewWhatsmeowLogger(logger utils.LogrusEntry) *WhatsmeowLogger {
	return &WhatsmeowLogger{
		logger: logger,
		name:   "whatsmeow",
	}
}

// Debugf implementasi waLog.Logger
func (l *WhatsmeowLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Infof implementasi waLog.Logger
func (l *WhatsmeowLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf implementasi waLog.Logger
func (l *WhatsmeowLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf implementasi waLog.Logger
func (l *WhatsmeowLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Sub implementasi waLog.Logger
func (l *WhatsmeowLogger) Sub(module string) waLog.Logger {
	return &WhatsmeowLogger{
		logger: l.logger.WithField("sub", module),
		name:   module,
	}
}
