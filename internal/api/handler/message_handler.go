package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// Konstanta untuk pesan broadcast
const (
	MsgBroadcastSuccess   = "Broadcast berhasil diproses"
	MsgMinTarget          = "Minimal harus ada 2 target penerima untuk broadcast"
	MsgMaxTarget          = "Maksimal hanya 16 target penerima untuk broadcast"
	MsgEmptyMessage       = "Pesan tidak boleh kosong"
	DefaultBroadcastDelay = 1000
	MaxBroadcastDelay     = 5000
)

// MessageHandler menangani endpoint pesan API
type MessageHandler struct {
	BaseHandler
}

// NewMessageHandler membuat instance baru MessageHandler
func NewMessageHandler(whatsClient *client.Client) *MessageHandler {
	return &MessageHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-message"),
	}
}

// SendGroup mengirim pesan ke grup
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	var req model.GroupMessageRequest

	// Validasi request DENGAN LANGSUNG RETURN jika gagal
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	// Validasi required fields secara eksplisit dengan early return
	if req.GroupID == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "groupID"), nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "message"), nil, fiber.StatusBadRequest)
	}

	// Validasi tambahan untuk group ID
	if !utils.ValidateGroupID(req.GroupID) {
		return h.SendError(c, constants.MsgInvalidGroupID, nil, fiber.StatusBadRequest)
	}

	// Kirim pesan hanya jika validasi berhasil
	jid := client.ParseGroupID(req.GroupID)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan grup")
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan grup berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendGroupSuccess,
		jid.String(),
		"group"))
}

// SendPersonal mengirim pesan ke nomor personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, constants.MsgSendFailure+": "+constants.MsgNotConnected) {
		return nil
	}

	var req model.PersonalMessageRequest

	// Validasi request dengan parsing langsung
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body")
		return h.SendError(c, constants.MsgInvalidRequest, err, fiber.StatusBadRequest)
	}

	// Validasi required fields dengan early return
	if req.PhoneNumber == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "phoneNumber"), nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, fmt.Sprintf(constants.MsgMissingField, "message"), nil, fiber.StatusBadRequest)
	}

	// Validasi tambahan untuk nomor telepon
	if !utils.ValidatePhoneNumber(req.PhoneNumber) {
		return h.SendError(c, constants.MsgInvalidPhoneNumber, nil, fiber.StatusBadRequest)
	}

	// Tambahkan context timeout ke request fiber
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()
	c.SetUserContext(ctx)

	// Kirim pesan hanya jika validasi berhasil
	jid := client.ParsePhoneNumber(req.PhoneNumber)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.Logger.Error("Timeout saat mengirim pesan personal")
			return h.SendError(c, constants.MsgTimeoutError, err, fiber.StatusGatewayTimeout)
		}
		h.Logger.WithError(err).Error("Gagal mengirim pesan personal")
		return h.SendError(c, constants.MsgSendFailure, err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan personal berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		constants.MsgSendSuccess,
		jid.String(),
		"personal"))
}

// Broadcast mengirim pesan ke banyak nomor/grup sekaligus
func (h *MessageHandler) Broadcast(c *fiber.Ctx) error {
	// Gunakan metode standar untuk memeriksa koneksi
	if !h.CheckWhatsAppConnection(c, "Gagal mengirim pesan broadcast: WhatsApp sedang tidak terhubung") {
		return nil // Respons sudah dikirim oleh CheckWhatsAppConnection
	}

	// Debug lebih detail untuk melihat struktur JSON asli
	bodyBytes := c.Body()
	h.Logger.Debug(fmt.Sprintf("Request body raw: %s", string(bodyBytes)))

	// Parse request
	var req model.BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body broadcast")
		return h.SendError(c, "Format request tidak valid", err, fiber.StatusBadRequest)
	}

	// Debugging
	h.Logger.Debug(fmt.Sprintf("Personal numbers raw: %+v", req.PersonalNumbers))
	h.Logger.Debug(fmt.Sprintf("Group IDs raw: %+v", req.GroupIDs))

	// Proses dan deduplikasi personalNumbers
	cleanedPersonalNumbers, dupPersonalCount := h.cleanAndDedupPersonalNumbers(req.PersonalNumbers)

	// Proses dan deduplikasi groupIDs
	cleanedGroupIDs, dupGroupCount := h.cleanAndDedupGroupIDs(req.GroupIDs)

	// Log hasil deduplikasi
	if dupPersonalCount > 0 || dupGroupCount > 0 {
		h.Logger.Info("Ditemukan dan dilewati target duplikat", utils.Fields{
			"duplicate_personal": dupPersonalCount,
			"duplicate_groups":   dupGroupCount,
		})
	}

	// Update request dengan data yang sudah dibersihkan
	req.PersonalNumbers = cleanedPersonalNumbers
	req.GroupIDs = cleanedGroupIDs

	// Tambahkan log untuk verifikasi hasil cleaning
	h.Logger.WithFields(utils.Fields{
		"cleaned_personal_count": len(cleanedPersonalNumbers),
		"cleaned_group_count":    len(cleanedGroupIDs),
		"cleaned_personal":       cleanedPersonalNumbers,
		"cleaned_groups":         cleanedGroupIDs,
	}).Debug("Hasil pembersihan input")

	// Hitung ulang total target
	totalTargets := len(req.PersonalNumbers) + len(req.GroupIDs)

	// Validasi jumlah minimum target
	if totalTargets < 2 {
		return h.SendError(c, MsgMinTarget, nil, fiber.StatusBadRequest)
	}

	// Validasi jumlah maksimum target
	if totalTargets > 16 {
		return h.SendError(c, MsgMaxTarget, nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, MsgEmptyMessage, nil, fiber.StatusBadRequest)
	}

	// Sesuaikan delay jika tidak dalam rentang yang wajar
	if req.DelayMs <= 0 {
		req.DelayMs = DefaultBroadcastDelay // Default 1 detik
	} else if req.DelayMs > MaxBroadcastDelay {
		req.DelayMs = MaxBroadcastDelay // Maksimal 5 detik
	}

	// Catat waktu mulai untuk menghitung durasi proses
	startTime := time.Now()

	// Proses broadcast
	h.Logger.WithFields(utils.Fields{
		"personal_count":       len(req.PersonalNumbers),
		"group_count":          len(req.GroupIDs),
		"delay_ms":             req.DelayMs,
		"personal_numbers_raw": req.PersonalNumbers, // Log array mentah untuk debugging
	}).Info("Memulai pengiriman broadcast")

	// Cetak setiap nomor secara individual untuk memastikan parsing benar
	for i, num := range req.PersonalNumbers {
		h.Logger.Debug(fmt.Sprintf("Personal number #%d: %s", i+1, num))
	}

	// Lakukan broadcast dengan data yang sudah bersih
	results := h.WhatsApp.BroadcastMessage(
		req.PersonalNumbers,
		req.GroupIDs,
		req.Message,
		req.DelayMs,
	)

	// Hitung waktu pemrosesan
	processingTime := time.Since(startTime).Milliseconds()

	// Hitung statistik hasil
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	// Log hasil
	h.Logger.WithFields(utils.Fields{
		"total":           len(results),
		"success":         successCount,
		"failed":          len(results) - successCount,
		"processing_time": processingTime,
	}).Info("Broadcast selesai")

	// Buat respons
	response := model.NewBroadcastResponse(
		MsgBroadcastSuccess,
		results,
		processingTime,
	)

	return h.SendSuccess(c, response)
}

// cleanAndDedupPersonalNumbers membersihkan dan mendeduplikasi nomor personal
func (h *MessageHandler) cleanAndDedupPersonalNumbers(numbers []string) ([]string, int) {
	cleanedPersonalNumbers := make([]string, 0)
	seenPersonalNumbers := make(map[string]bool)
	dupPersonalCount := 0 // Ubah: Mulai dari 0 dan tambah saat menemukan duplikat

	for _, num := range numbers {
		// Lompati nilai kosong/array kosong
		if num == "[]" || num == "[ ]" || num == "" {
			continue
		}

		// Tangani kasus array dalam string
		if strings.HasPrefix(num, "[") && strings.HasSuffix(num, "]") {
			h.Logger.Debug(fmt.Sprintf("Memproses array dalam string: %s", num))

			// Ekstrak nilai-nilai dalam array
			innerArray := extractArrayValues(num)
			// Ubah: Proses hasil ekstraksi dengan deduplikasi
			for _, innerNum := range innerArray {
				if innerNum == "" || !utils.ValidatePhoneNumber(innerNum) {
					continue
				}

				// Normalisasi untuk deduplikasi
				normalizedNum := utils.FormatPhoneNumber(innerNum)

				if !seenPersonalNumbers[normalizedNum] {
					seenPersonalNumbers[normalizedNum] = true
					cleanedPersonalNumbers = append(cleanedPersonalNumbers, innerNum)
				} else {
					dupPersonalCount++ // Tambahkan counter duplikat
					h.Logger.Info("Melewati nomor duplikat", utils.Fields{
						"number": innerNum,
					})
				}
			}
			continue
		}

		// Nomor normal
		num = strings.TrimSpace(num)
		if num == "" || !utils.ValidatePhoneNumber(num) {
			continue
		}

		// Deduplikasi nomor normal
		normalizedNum := utils.FormatPhoneNumber(num)
		if !seenPersonalNumbers[normalizedNum] {
			seenPersonalNumbers[normalizedNum] = true
			cleanedPersonalNumbers = append(cleanedPersonalNumbers, num)
		} else {
			dupPersonalCount++ // Tambahkan counter duplikat
			h.Logger.Info("Melewati nomor duplikat", utils.Fields{
				"number": num,
			})
		}
	}

	return cleanedPersonalNumbers, dupPersonalCount
}

// cleanAndDedupGroupIDs membersihkan dan mendeduplikasi ID grup
func (h *MessageHandler) cleanAndDedupGroupIDs(ids []string) ([]string, int) {
	cleanedGroupIDs := make([]string, 0)
	seenGroupIDs := make(map[string]bool)
	dupGroupCount := 0 // Ubah: Mulai dari 0 dan tambah saat menemukan duplikat

	for _, id := range ids {
		// Lompati nilai kosong/array kosong
		if id == "[]" || id == "[ ]" || id == "" {
			continue
		}

		// Tangani kasus array dalam string
		if strings.HasPrefix(id, "[") && strings.HasSuffix(id, "]") {
			h.Logger.Debug(fmt.Sprintf("Memproses array dalam string: %s", id))

			// Ekstrak nilai-nilai dalam array
			innerArray := extractArrayValues(id)
			// Ubah: Proses hasil ekstraksi dengan deduplikasi
			for _, innerID := range innerArray {
				if innerID == "" || !utils.ValidateGroupID(innerID) {
					continue
				}

				// Normalisasi untuk deduplikasi
				normalizedID := normalizeGroupID(innerID)

				if !seenGroupIDs[normalizedID] {
					seenGroupIDs[normalizedID] = true
					cleanedGroupIDs = append(cleanedGroupIDs, innerID)
				} else {
					dupGroupCount++ // Tambahkan counter duplikat
					h.Logger.Info("Melewati ID grup duplikat", utils.Fields{
						"group_id": innerID,
					})
				}
			}
			continue
		}

		// Group ID normal
		id = strings.TrimSpace(id)
		if id == "" || !utils.ValidateGroupID(id) {
			continue
		}

		// Deduplikasi group ID
		normalizedID := normalizeGroupID(id)
		if !seenGroupIDs[normalizedID] {
			seenGroupIDs[normalizedID] = true
			cleanedGroupIDs = append(cleanedGroupIDs, id)
		} else {
			dupGroupCount++ // Tambahkan counter duplikat
			h.Logger.Info("Melewati ID grup duplikat", utils.Fields{
				"group_id": id,
			})
		}
	}

	return cleanedGroupIDs, dupGroupCount
}

// extractArrayValues mengekstrak nilai-nilai dari string array JSON
func extractArrayValues(arrayStr string) []string {
	// Hapus tanda bracket di awal dan akhir
	content := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(arrayStr, "]"), "["))
	if content == "" {
		return []string{}
	}

	// Split berdasarkan koma tetapi hitung tanda kutip
	var result []string
	var currentValue strings.Builder
	inQuotes := false

	for _, char := range content {
		switch char {
		case '"', '\'':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				// Koma di luar tanda kutip = pemisah
				val := strings.TrimSpace(currentValue.String())
				// Bersihkan tanda kutip
				val = strings.Trim(val, `"'`)
				result = append(result, val)
				currentValue.Reset()
				continue
			}
		}
		currentValue.WriteRune(char)
	}

	// Tambahkan nilai terakhir jika ada
	if currentValue.Len() > 0 {
		val := strings.TrimSpace(currentValue.String())
		val = strings.Trim(val, `"'`)
		result = append(result, val)
	}

	return result
}

// normalizeGroupID adalah helper untuk memastikan ID grup dinormalisasi secara konsisten
func normalizeGroupID(id string) string {
	// Hapus @g.us jika ada
	id = strings.Split(id, "@")[0]

	// Hapus semua karakter non-digit
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, id)
}
