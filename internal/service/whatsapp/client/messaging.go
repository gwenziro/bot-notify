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

// SendMessage mengirim pesan teks ke nomor atau grup tertentu
func (c *Client) SendMessage(recipient types.JID, message string) (time.Time, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return time.Time{}, errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.WithFields(utils.Fields{
		"to":             recipient.String(),
		"message_length": len(message),
	}).Info("Mengirim pesan")

	// Update aktivitas
	c.UpdateLastActivity()

	// Tambahkan timeout 10 detik untuk operasi pengiriman pesan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Catat waktu pengiriman sebenarnya
	sendTime := time.Now()

	// Kirim pesan dengan context timeout
	_, err := c.waClient.SendMessage(ctx, recipient, &waE2E.Message{
		Conversation: &message,
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return time.Time{}, fmt.Errorf("timeout saat mengirim pesan: operasi melebihi 10 detik")
		}
		return time.Time{}, fmt.Errorf("gagal mengirim pesan: %w", err)
	}

	return sendTime, nil
}

// SendFormattedMessage mengirim pesan dengan format khusus (bold, italic, dll)
func (c *Client) SendFormattedMessage(recipient types.JID, message string) error {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.WithFields(utils.Fields{
		"to":             recipient.String(),
		"message_length": len(message),
		"type":           "formatted",
	}).Info("Mengirim pesan terformat")

	// Update aktivitas
	c.UpdateLastActivity()

	// Konversi ke ExtendedTextMessage untuk dukungan format
	_, err := c.waClient.SendMessage(context.Background(), recipient, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &message,
			// Bisa ditambahkan opsi pemformatan lainnya
		},
	})

	if err != nil {
		return fmt.Errorf("gagal mengirim pesan terformat: %w", err)
	}

	return nil
}

// BroadcastMessage mengirim pesan ke beberapa target sekaligus
func (c *Client) BroadcastMessage(personalNumbers []string, groupIDs []string, message string, delayMs int) []model.BroadcastResult {
	results := make([]model.BroadcastResult, 0, len(personalNumbers)+len(groupIDs))

	// Debug log yang lebih jelas (tanpa menggabungkan array menjadi string)
	c.logger.WithFields(utils.Fields{
		"personal_count":     len(personalNumbers),
		"group_count":        len(groupIDs),
		"message_length":     len(message),
		"first_personal_num": getFirstOrEmpty(personalNumbers),
		"first_group_id":     getFirstOrEmpty(groupIDs),
	}).Debug("Menerima permintaan broadcast")

	// Kirim ke nomor personal (satu per satu)
	for i, number := range personalNumbers {
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
		jid := ParsePhoneNumber(number)

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid.String(),
			"type": "personal",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		_, err := c.SendMessage(jid, message)

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
			c.logger.Info("Berhasil mengirim pesan broadcast personal", utils.Fields{
				"target": number,
			})
		}

		results = append(results, result)

		// Delay untuk mencegah throttling
		if delayMs > 0 && len(personalNumbers) > 1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	// Kirim ke grup
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

		// Log grup yang sedang diproses untuk debugging
		c.logger.WithFields(utils.Fields{
			"group_id": groupID,
			"type":     "group",
		}).Info("Memproses target broadcast grup")

		// Parse ID grup ke JID
		jid := ParseGroupID(groupID)

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid.String(),
			"type": "group",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		_, err := c.SendMessage(jid, message)

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

// Helper function untuk mendapatkan elemen pertama array atau string kosong
func getFirstOrEmpty(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}
