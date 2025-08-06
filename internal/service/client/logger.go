package client

import (
	"fmt"
	"strings"

	"github.com/gwenziro/bot-notify/internal/utils"
	log "go.mau.fi/whatsmeow/util/log"
)

// WhatsmeowLogger merupakan implementasi logger untuk whatsmeow
type WhatsmeowLogger struct {
	logger utils.LogrusEntry
}

// NewWhatsmeowLogger membuat logger baru untuk whatsmeow
func NewWhatsmeowLogger(logger utils.LogrusEntry) *WhatsmeowLogger {
	return &WhatsmeowLogger{
		logger: logger,
	}
}

// shouldFilterMessage menentukan apakah pesan log harus difilter
func (l *WhatsmeowLogger) shouldFilterMessage(msg string) bool {
	// Filter pesan yang berkaitan dengan status/story WhatsApp (biasanya mengandung "@broadcast")
	if strings.Contains(msg, "@broadcast") {
		return true
	}

	// Filter error decrypting khusus untuk status broadcast
	if strings.Contains(msg, "Error decrypting message") && strings.Contains(msg, "status@broadcast") {
		return true
	}

	// Filter error terkait sender key untuk status broadcast
	if strings.Contains(msg, "failed to decrypt group message: no sender key state for key ID") &&
		strings.Contains(msg, "status@broadcast") {
		return true
	}

	return false
}

// Debugf log dengan level Debug
func (l *WhatsmeowLogger) Debugf(msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	if !l.shouldFilterMessage(formattedMsg) {
		l.logger.Debug(formattedMsg)
	}
}

// Infof log dengan level Info
func (l *WhatsmeowLogger) Infof(msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	if !l.shouldFilterMessage(formattedMsg) {
		l.logger.Info(formattedMsg)
	}
}

// Warnf log dengan level Warn
func (l *WhatsmeowLogger) Warnf(msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	if !l.shouldFilterMessage(formattedMsg) {
		l.logger.Warn(formattedMsg)
	}
}

// Errorf log dengan level Error
func (l *WhatsmeowLogger) Errorf(msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	if !l.shouldFilterMessage(formattedMsg) {
		l.logger.Error(formattedMsg)
	}
}

// Sub implementasikan interface SubLogger
func (l *WhatsmeowLogger) Sub(module string) log.Logger {
	return &WhatsmeowLogger{
		logger: l.logger.WithField("whatsmeow-module", module),
	}
}
