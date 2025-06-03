package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// InternalBroadcastResult adalah struktur internal untuk hasil broadcast per target
type InternalBroadcastResult struct {
	Target   string
	Type     string // "personal" atau "group"
	Success  bool
	ErrorMsg string
	SentTime time.Time // Waktu pengiriman aktual jika berhasil
}

// SendMessage mengirim pesan teks ke nomor atau grup tertentu
func (c *Client) SendMessage(recipient types.JID, message string) (time.Time, error) {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return time.Time{}, errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.WithFields(utils.Fields{
		"to":             recipient.String(),
		"message_length": len(message),
	}).Info("Mengirim pesan")

	// Hapus c.UpdateLastActivity()

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

	// Hapus c.UpdateLastActivity()

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
// Mengembalikan slice dari InternalBroadcastResult dan waktu pengiriman terakhir yang berhasil.
func (c *Client) BroadcastMessage(personalNumbers []string, groupIDs []string, message string, delayMs int) ([]InternalBroadcastResult, time.Time) {
	results := make([]InternalBroadcastResult, 0, len(personalNumbers)+len(groupIDs))
	var lastSentTime time.Time

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
			results = append(results, InternalBroadcastResult{
				Target:   number,
				Type:     "personal",
				Success:  false,
				ErrorMsg: "Nomor telepon tidak valid",
			})
			continue
		}

		// Parse nomor telepon ke JID
		phoneNumber := utils.FormatPhoneNumber(number)
		jid, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
		if err != nil {
			c.logger.Warn("Gagal parsing JID", utils.Fields{"number": number, "error": err.Error()})
			results = append(results, InternalBroadcastResult{
				Target:   number,
				Type:     "personal",
				Success:  false,
				ErrorMsg: "Gagal parsing JID: " + err.Error(),
			})
			continue
		}

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid,
			"type": "personal",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		sentTime, err := c.SendMessage(jid, message)

		// Update lastSentTime jika pengiriman berhasil
		if err == nil {
			lastSentTime = sentTime
		}

		// Catat hasil
		result := InternalBroadcastResult{
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
			result.SentTime = sentTime
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
	for i, rawGroupID := range groupIDs {
		// Lewati jika ID grup kosong
		if rawGroupID == "" {
			continue
		}

		// Deteksi apakah format array dengan kurung siku [id1, id2, ...]
		if strings.HasPrefix(rawGroupID, "[") && strings.HasSuffix(rawGroupID, "]") {
			// Ini adalah format array, proses masing-masing ID
			c.logger.Debug("Mendeteksi grup ID dalam format array", utils.Fields{
				"raw_input": rawGroupID,
			})

			// Ekstrak konten di dalam kurung siku
			content := rawGroupID[1 : len(rawGroupID)-1]

			// Split berdasarkan koma untuk mendapatkan setiap ID
			groupIDItems := strings.Split(content, ",")

			// Tambahkan setiap ID grup ke slice groupIDs untuk diproses
			for _, groupIDItem := range groupIDItems {
				trimmedID := strings.TrimSpace(groupIDItem)
				if trimmedID != "" {
					// Rekursif: panggil BroadcastMessage lagi dengan ID yang sudah diekstrak
					// Gunakan delay yang sama untuk konsistensi
					subResults, subSentTime := c.BroadcastMessage(
						[]string{},          // Tidak ada nomor personal
						[]string{trimmedID}, // Hanya satu ID grup yang diproses
						message,
						delayMs,
					)

					// Tambahkan hasil ke hasil utama
					results = append(results, subResults...)

					// Update lastSentTime jika ada pengiriman yang berhasil
					if !subSentTime.IsZero() {
						lastSentTime = subSentTime
					}
				}
			}

			// Skip pemrosesan item saat ini karena sudah diproses dalam loop di atas
			continue
		}

		// Debug log untuk melihat format ID grup yang diterima
		c.logger.Debug("Memproses ID grup untuk broadcast", utils.Fields{
			"raw_group_id": rawGroupID,
			"index":        i,
		})

		// Validasi ID grup sekaligus konversi langsung ke JID
		jid := utils.ParseGroupID(rawGroupID)

		// Jika invalid, ParseGroupID akan mengembalikan JID dengan user part "invalid"
		if jid.User == "invalid" {
			c.logger.Warn("ID grup tidak valid, dilewati", utils.Fields{
				"group_id": rawGroupID,
			})
			results = append(results, InternalBroadcastResult{
				Target:   rawGroupID,
				Type:     "group",
				Success:  false,
				ErrorMsg: "ID grup tidak valid",
			})
			continue
		}

		// Catat mulai pengiriman
		c.logger.WithFields(utils.Fields{
			"to":   jid,
			"type": "group",
		}).Info("Mengirim pesan broadcast")

		// Kirim pesan dan ambil timestamp pengiriman
		sentTime, err := c.SendMessage(jid, message)

		// Update lastSentTime jika pengiriman berhasil
		if err == nil {
			lastSentTime = sentTime
		}

		// Catat hasil
		result := InternalBroadcastResult{
			Target:  rawGroupID,
			Type:    "group",
			Success: err == nil,
		}

		if err != nil {
			result.ErrorMsg = err.Error()
			c.logger.WithError(err).Warn("Gagal mengirim pesan broadcast", utils.Fields{
				"target": rawGroupID,
				"type":   "group",
			})
		} else {
			// Tambahkan waktu pengiriman jika berhasil
			result.SentTime = sentTime
			c.logger.Info("Berhasil mengirim pesan broadcast grup", utils.Fields{
				"target": rawGroupID,
			})
		}

		results = append(results, result)

		// Delay untuk mencegah throttling (kecuali di iterasi terakhir)
		if delayMs > 0 && i < len(groupIDs)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	// Jika tidak ada pengiriman yang berhasil, gunakan waktu sekarang
	if lastSentTime.IsZero() {
		lastSentTime = time.Now()
	}

	return results, lastSentTime
}

// Helper function untuk mendapatkan elemen pertama array atau string kosong
func getFirstOrEmpty(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}
