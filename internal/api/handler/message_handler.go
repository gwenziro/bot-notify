package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
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

// SendPersonal mengirim pesan ke nomor personal
func (h *MessageHandler) SendPersonal(c *fiber.Ctx) error {
	// Dapatkan status koneksi terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas
	if !state.IsConnected {
		h.Logger.Info("Permintaan kirim pesan personal saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
	}

	var req model.PersonalMessageRequest

	// Validasi request
	if err := h.ValidateRequest(c, &req, map[string]func() string{
		"phoneNumber": func() string { return req.PhoneNumber },
		"message":     func() string { return req.Message },
	}); err != nil {
		return err
	}

	// Kirim pesan
	jid := client.ParsePhoneNumber(req.PhoneNumber)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan personal")
		return h.SendError(c, "Gagal mengirim pesan", err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan personal berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim!",
		jid.String(),
		"personal"))
}

// SendGroup mengirim pesan ke grup
func (h *MessageHandler) SendGroup(c *fiber.Ctx) error {
	// Dapatkan status koneksi terlebih dahulu
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas
	if !state.IsConnected {
		h.Logger.Info("Permintaan kirim pesan grup saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
	}

	var req model.GroupMessageRequest

	// Validasi request
	if err := h.ValidateRequest(c, &req, map[string]func() string{
		"groupID": func() string { return req.GroupID },
		"message": func() string { return req.Message },
	}); err != nil {
		return err
	}

	// Kirim pesan
	jid := client.ParseGroupID(req.GroupID)
	if err := h.WhatsApp.SendMessage(jid, req.Message); err != nil {
		h.Logger.WithError(err).Error("Gagal mengirim pesan grup")
		return h.SendError(c, "Gagal mengirim pesan", err, fiber.StatusInternalServerError)
	}

	h.Logger.Info("Pesan grup berhasil dikirim")
	return h.SendSuccess(c, model.NewMessageResponse(
		"Notifikasi WhatsApp terkirim ke grup!",
		jid.String(),
		"group"))
}

// Broadcast mengirim pesan ke banyak nomor/grup sekaligus
func (h *MessageHandler) Broadcast(c *fiber.Ctx) error {
	// Periksa koneksi WhatsApp
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	// Jika tidak terhubung, kembalikan error yang jelas
	if !state.IsConnected {
		h.Logger.Info("Permintaan broadcast saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan broadcast: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
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

	// Proses personalNumbers dengan deduplikasi yang benar
	cleanedPersonalNumbers := make([]string, 0)
	seenPersonalNumbers := make(map[string]bool)
	dupPersonalCount := 0 // Ubah: Mulai dari 0 dan tambah saat menemukan duplikat

	for _, num := range req.PersonalNumbers {
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

	// Proses groupIDs dengan deduplikasi yang benar
	cleanedGroupIDs := make([]string, 0)
	seenGroupIDs := make(map[string]bool)
	dupGroupCount := 0 // Ubah: Mulai dari 0 dan tambah saat menemukan duplikat

	for _, id := range req.GroupIDs {
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

	// Log jumlah duplikat yang ditemukan dan diproses
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
		return h.SendError(c, "Minimal harus ada 2 target penerima untuk broadcast", nil, fiber.StatusBadRequest)
	}

	// Validasi jumlah maksimum target
	if totalTargets > 16 {
		return h.SendError(c, "Maksimal hanya 16 target penerima untuk broadcast", nil, fiber.StatusBadRequest)
	}

	if req.Message == "" {
		return h.SendError(c, "Pesan tidak boleh kosong", nil, fiber.StatusBadRequest)
	}

	// Sesuaikan delay jika tidak dalam rentang yang wajar
	if req.DelayMs <= 0 {
		req.DelayMs = 1000 // Default 1 detik
	} else if req.DelayMs > 5000 {
		req.DelayMs = 5000 // Maksimal 5 detik
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
		"Broadcast berhasil diproses",
		results,
		processingTime,
	)

	return h.SendSuccess(c, response)
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
