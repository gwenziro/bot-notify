package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// sendBaseMessage adalah fungsi dasar untuk mengirim berbagai jenis pesan
// Parameters:
// - recipient: JID penerima
// - content: konten pesan (map ke waE2E.Message)
// - timeout: durasi timeout dalam detik (0 untuk default 10 detik)
// Returns:
// - waktu pengiriman jika berhasil
// - error jika gagal
func (c *Client) sendBaseMessage(recipient types.JID, content *waE2E.Message, timeout time.Duration) (time.Time, error) {
	if err := c.validateConnection(); err != nil {
		return time.Time{}, err
	}

	c.logger.WithFields(utils.Fields{
		"to":   recipient.String(),
		"type": getMessageType(content),
	}).Info("Mengirim pesan")

	// Update aktivitas
	c.UpdateLastActivity()

	// Buat context dengan timeout
	ctx, cancel := c.createTimeoutContext(timeout)
	defer cancel()

	// Catat waktu pengiriman sebenarnya
	sendTime := time.Now()

	// Kirim pesan dengan context timeout
	_, err := c.waClient.SendMessage(ctx, recipient, content)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return time.Time{}, fmt.Errorf("timeout saat mengirim pesan: operasi melebihi batas waktu")
		}
		return time.Time{}, fmt.Errorf("gagal mengirim pesan: %w", err)
	}

	return sendTime, nil
}

// getMessageType menentukan tipe pesan untuk logging
func getMessageType(msg *waE2E.Message) string {
	if msg.Conversation != nil {
		return "text"
	} else if msg.ExtendedTextMessage != nil {
		return "formatted"
	} else {
		return "other"
	}
}

// SendMessage mengirim pesan teks ke nomor atau grup tertentu
// Returns:
// - waktu pengiriman jika berhasil
// - error jika gagal
func (c *Client) SendMessage(recipient types.JID, message string) (time.Time, error) {
	return c.sendBaseMessage(recipient, &waE2E.Message{
		Conversation: &message,
	}, 10*time.Second)
}

// SendFormattedMessage mengirim pesan dengan format khusus (bold, italic, dll)
// Returns:
// - nil jika berhasil
// - error jika gagal
func (c *Client) SendFormattedMessage(recipient types.JID, message string) error {
	_, err := c.sendBaseMessage(recipient, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &message,
			// Bisa ditambahkan opsi pemformatan lainnya
		},
	}, 10*time.Second)
	return err
}

// BroadcastMessage mengirim pesan ke beberapa target sekaligus
// Parameters:
// - personalNumbers: daftar nomor telepon target
// - groupIDs: daftar ID grup target
// - message: pesan yang akan dikirim
// - delayMs: delay antara pengiriman (millisecond)
// Returns:
// - hasil broadcast untuk setiap target
// - waktu pengiriman terakhir
func (c *Client) BroadcastMessage(personalNumbers []string, groupIDs []string, message string, delayMs int) ([]model.BroadcastResult, time.Time) {
	results := make([]model.BroadcastResult, 0, len(personalNumbers)+len(groupIDs))
	var lastSentTime time.Time

	// Debug log yang lebih jelas
	c.logger.WithFields(utils.Fields{
		"personal_count":     len(personalNumbers),
		"group_count":        len(groupIDs),
		"message_length":     len(message),
		"first_personal_num": utils.GetFirstOrEmpty(personalNumbers),
		"first_group_id":     utils.GetFirstOrEmpty(groupIDs),
	}).Debug("Menerima permintaan broadcast")

	// Kirim ke nomor personal (satu per satu)
	results = append(results, c.broadcastToPersonal(personalNumbers, message, delayMs, &lastSentTime)...)

	// Kirim ke grup
	results = append(results, c.broadcastToGroups(groupIDs, message, delayMs, &lastSentTime)...)

	// Jika tidak ada pengiriman yang berhasil, gunakan waktu sekarang
	if lastSentTime.IsZero() {
		lastSentTime = time.Now()
	}

	return results, lastSentTime
}

// broadcastToPersonal mengirim pesan ke daftar nomor personal
func (c *Client) broadcastToPersonal(numbers []string, message string, delayMs int, lastSentTime *time.Time) []model.BroadcastResult {
	results := make([]model.BroadcastResult, 0, len(numbers))

	for i, number := range numbers {
		// Log dengan index untuk clarity
		c.logger.WithFields(utils.Fields{
			"index":  i,
			"number": number,
			"type":   "personal",
		}).Info("Memproses target broadcast personal")

		// Validasi nomor telepon
		if !utils.ValidatePhoneNumber(number) {
			c.logger.Warn("Nomor telepon tidak valid, dilewati", utils.Fields{"number": number})
			results = append(results, model.BroadcastResult{
				Target:   number,
				Type:     "personal",
				Success:  false,
				ErrorMsg: "Nomor telepon tidak valid",
			})
			continue
		}

		// Parse nomor telepon ke JID
		jidString := utils.FormatWhatsAppNumber(number)
		jid, parseErr := types.ParseJID(jidString)
		if parseErr != nil {
			c.logger.WithError(parseErr).Warn("Gagal parsing JID", utils.Fields{"number": number})
			results = append(results, model.BroadcastResult{
				Target:   number,
				Type:     "personal",
				Success:  false,
				ErrorMsg: "Gagal parsing JID: " + parseErr.Error(),
			})
			continue
		}

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid.String(),
			"type": "personal",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		sentTime, err := c.SendMessage(jid, message)

		// Update lastSentTime jika pengiriman berhasil
		if err == nil && lastSentTime != nil {
			*lastSentTime = sentTime
		}

		// Catat hasil
		result := model.BroadcastResult{
			Target:  number,
			Type:    "personal",
			Success: err == nil,
		}

		if err != nil {
			result.ErrorMsg = err.Error()
			c.logger.WithError(err).Warn("Gagal mengirim pesan broadcast", utils.Fields{
				"target": number,
				"type":   "personal",
			})
		} else {
			// Tambahkan waktu pengiriman jika berhasil
			result.SentTime = utils.FormatTimeIndonesia(&sentTime)
			c.logger.Info("Berhasil mengirim pesan broadcast personal", utils.Fields{
				"target": number,
			})
		}

		results = append(results, result)

		// Delay untuk mencegah throttling
		if delayMs > 0 && i < len(numbers)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	return results
}

// broadcastToGroups mengirim pesan ke daftar grup
func (c *Client) broadcastToGroups(groupIDs []string, message string, delayMs int, lastSentTime *time.Time) []model.BroadcastResult {
	results := make([]model.BroadcastResult, 0, len(groupIDs))

	for i, groupID := range groupIDs {
		// Lewati jika ID grup kosong
		if groupID == "" {
			continue
		}

		// Validasi ID grup
		if !utils.ValidateGroupID(groupID) {
			c.logger.Warn("ID grup tidak valid, dilewati", utils.Fields{"group_id": groupID})
			results = append(results, model.BroadcastResult{
				Target:   groupID,
				Type:     "group",
				Success:  false,
				ErrorMsg: "ID grup tidak valid",
			})
			continue
		}

		// Parse ID grup ke JID
		jidString := utils.FormatGroupID(groupID)
		jid, parseErr := types.ParseJID(jidString)
		if parseErr != nil {
			c.logger.WithError(parseErr).Warn("Gagal parsing JID grup", utils.Fields{"group_id": groupID})
			results = append(results, model.BroadcastResult{
				Target:   groupID,
				Type:     "group",
				Success:  false,
				ErrorMsg: "Gagal parsing JID: " + parseErr.Error(),
			})
			continue
		}

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid.String(),
			"type": "group",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		sentTime, err := c.SendMessage(jid, message)

		// Update lastSentTime jika pengiriman berhasil
		if err == nil && lastSentTime != nil {
			*lastSentTime = sentTime
		}

		// Catat hasil
		result := model.BroadcastResult{
			Target:  groupID,
			Type:    "group",
			Success: err == nil,
		}

		if err != nil {
			result.ErrorMsg = err.Error()
			c.logger.WithError(err).Warn("Gagal mengirim pesan broadcast", utils.Fields{
				"target": groupID,
				"type":   "group",
			})
		} else {
			// Tambahkan waktu pengiriman jika berhasil
			result.SentTime = utils.FormatTimeIndonesia(&sentTime)
			c.logger.Info("Berhasil mengirim pesan broadcast grup", utils.Fields{
				"target": groupID,
			})
		}

		results = append(results, result)

		// Delay untuk mencegah throttling (kecuali di iterasi terakhir)
		if delayMs > 0 && i < len(groupIDs)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	return results
}
