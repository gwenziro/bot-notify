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
	// Validasi koneksi
	if err := h.validateConnection(c); err != nil {
		return err
	}

	// Parse dan validasi request
	req, err := h.parseBroadcastRequest(c)
	if err != nil {
		return err
	}

	// Proses dan deduplikasi nomor personal
	cleanedPersonalNumbers, dupPersonalCount := h.processPersonalNumbers(req.PersonalNumbers)
	req.PersonalNumbers = cleanedPersonalNumbers

	// Proses dan deduplikasi ID grup
	cleanedGroupIDs, dupGroupCount := h.processGroupIDs(req.GroupIDs)
	req.GroupIDs = cleanedGroupIDs

	// Log tentang deduplikasi
	h.logDeduplikasi(dupPersonalCount, dupGroupCount, cleanedPersonalNumbers, cleanedGroupIDs)

	// Validasi jumlah target
	if err := h.validateTargetCount(c, len(cleanedPersonalNumbers)+len(cleanedGroupIDs)); err != nil {
		return err
	}

	// Validasi pesan
	if req.Message == "" {
		return h.SendError(c, "Pesan tidak boleh kosong", nil, fiber.StatusBadRequest)
	}

	// Sesuaikan delay
	req.DelayMs = h.normalizeDelay(req.DelayMs)

	// Lakukan broadcast dan dapatkan hasilnya
	return h.executeBroadcast(c, req)
}

// validateConnection memeriksa apakah WhatsApp terhubung
func (h *MessageHandler) validateConnection(c *fiber.Ctx) error {
	state, err := h.WhatsApp.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, "Gagal mendapatkan status koneksi", err, fiber.StatusInternalServerError)
	}

	if !state.IsConnected {
		h.Logger.Info("Permintaan broadcast saat WhatsApp tidak terhubung")
		errorResp := model.NewMessageResponse("Gagal mengirim pesan broadcast: WhatsApp sedang tidak terhubung", "", "")
		errorResp.BaseResponse.Success = false
		return c.Status(fiber.StatusServiceUnavailable).JSON(errorResp)
	}

	return nil
}

// parseBroadcastRequest mem-parsing request broadcast dari body request
func (h *MessageHandler) parseBroadcastRequest(c *fiber.Ctx) (*model.BroadcastRequest, error) {
	// Debug lebih detail untuk melihat struktur JSON asli
	bodyBytes := c.Body()
	h.Logger.Debug(fmt.Sprintf("Request body raw: %s", string(bodyBytes)))

	var req model.BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.WithError(err).Error("Gagal parsing request body broadcast")
		return nil, h.SendError(c, "Format request tidak valid", err, fiber.StatusBadRequest)
	}

	h.Logger.Debug(fmt.Sprintf("Personal numbers raw: %+v", req.PersonalNumbers))
	h.Logger.Debug(fmt.Sprintf("Group IDs raw: %+v", req.GroupIDs))

	return &req, nil
}

// processPersonalNumbers memproses dan mendeduplikasi nomor personal
func (h *MessageHandler) processPersonalNumbers(numbers []string) ([]string, int) {
	cleanedNumbers := make([]string, 0)
	seenNumbers := make(map[string]bool)
	dupCount := 0

	for _, num := range numbers {
		// Lompati nilai kosong/array kosong
		if utils.IsEmptyArrayString(num) || num == "" {
			continue
		}

		// Tangani kasus array dalam string
		if utils.IsArrayString(num) {
			h.Logger.Debug(fmt.Sprintf("Memproses array dalam string: %s", num))

			// Ekstrak nilai-nilai dalam array
			innerArray := utils.ExtractArrayValues(num)
			for _, innerNum := range innerArray {
				if innerNum == "" || !utils.ValidatePhoneNumber(innerNum) {
					continue
				}

				// Deduplikasi dengan normalisasi
				normalizedNum := utils.FormatPhoneNumber(innerNum)
				if !seenNumbers[normalizedNum] {
					seenNumbers[normalizedNum] = true
					cleanedNumbers = append(cleanedNumbers, innerNum)
				} else {
					dupCount++
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

		// Deduplikasi
		normalizedNum := utils.FormatPhoneNumber(num)
		if !seenNumbers[normalizedNum] {
			seenNumbers[normalizedNum] = true
			cleanedNumbers = append(cleanedNumbers, num)
		} else {
			dupCount++
			h.Logger.Info("Melewati nomor duplikat", utils.Fields{
				"number": num,
			})
		}
	}

	return cleanedNumbers, dupCount
}

// processGroupIDs memproses dan mendeduplikasi ID grup
func (h *MessageHandler) processGroupIDs(groupIDs []string) ([]string, int) {
	cleanedIDs := make([]string, 0)
	seenIDs := make(map[string]bool)
	dupCount := 0

	for _, id := range groupIDs {
		// Lompati nilai kosong/array kosong
		if utils.IsEmptyArrayString(id) || id == "" {
			continue
		}

		// Tangani kasus array dalam string
		if utils.IsArrayString(id) {
			h.Logger.Debug(fmt.Sprintf("Memproses array dalam string: %s", id))

			// Ekstrak nilai-nilai dalam array
			innerArray := utils.ExtractArrayValues(id)
			for _, innerID := range innerArray {
				if innerID == "" || !utils.ValidateGroupID(innerID) {
					continue
				}

				// Deduplikasi dengan normalisasi
				normalizedID := h.normalizeGroupID(innerID)
				if !seenIDs[normalizedID] {
					seenIDs[normalizedID] = true
					cleanedIDs = append(cleanedIDs, innerID)
				} else {
					dupCount++
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

		// Deduplikasi
		normalizedID := h.normalizeGroupID(id)
		if !seenIDs[normalizedID] {
			seenIDs[normalizedID] = true
			cleanedIDs = append(cleanedIDs, id)
		} else {
			dupCount++
			h.Logger.Info("Melewati ID grup duplikat", utils.Fields{
				"group_id": id,
			})
		}
	}

	return cleanedIDs, dupCount
}

// normalizeGroupID adalah helper untuk memastikan ID grup dinormalisasi secara konsisten
func (h *MessageHandler) normalizeGroupID(id string) string {
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

// logDeduplikasi mencatat informasi tentang deduplikasi
func (h *MessageHandler) logDeduplikasi(dupPersonalCount, dupGroupCount int, cleanedPersonalNumbers, cleanedGroupIDs []string) {
	// Log jumlah duplikat
	if dupPersonalCount > 0 || dupGroupCount > 0 {
		h.Logger.Info("Ditemukan dan dilewati target duplikat", utils.Fields{
			"duplicate_personal": dupPersonalCount,
			"duplicate_groups":   dupGroupCount,
		})
	}

	// Log hasil pembersihan
	h.Logger.WithFields(utils.Fields{
		"cleaned_personal_count": len(cleanedPersonalNumbers),
		"cleaned_group_count":    len(cleanedGroupIDs),
		"cleaned_personal":       cleanedPersonalNumbers,
		"cleaned_groups":         cleanedGroupIDs,
	}).Debug("Hasil pembersihan input")
}

// validateTargetCount memvalidasi jumlah target untuk broadcast
func (h *MessageHandler) validateTargetCount(c *fiber.Ctx, totalTargets int) error {
	// Validasi jumlah minimum target
	if totalTargets < 2 {
		return h.SendError(c, "Minimal harus ada 2 target penerima untuk broadcast", nil, fiber.StatusBadRequest)
	}

	// Validasi jumlah maksimum target
	if totalTargets > 16 {
		return h.SendError(c, "Maksimal hanya 16 target penerima untuk broadcast", nil, fiber.StatusBadRequest)
	}

	return nil
}

// normalizeDelay menormalkan nilai delay ke rentang yang wajar
func (h *MessageHandler) normalizeDelay(delayMs int) int {
	if delayMs <= 0 {
		return 1000 // Default 1 detik
	} else if delayMs > 5000 {
		return 5000 // Maksimal 5 detik
	}
	return delayMs
}

// executeBroadcast melakukan proses broadcast dan mengirim hasilnya
func (h *MessageHandler) executeBroadcast(c *fiber.Ctx, req *model.BroadcastRequest) error {
	// Catat waktu mulai untuk menghitung durasi proses
	startTime := time.Now()

	// Log informasi broadcast
	h.Logger.WithFields(utils.Fields{
		"personal_count": len(req.PersonalNumbers),
		"group_count":    len(req.GroupIDs),
		"delay_ms":       req.DelayMs,
	}).Info("Memulai pengiriman broadcast")

	// Cetak setiap nomor untuk debugging
	for i, num := range req.PersonalNumbers {
		h.Logger.Debug(fmt.Sprintf("Personal number #%d: %s", i+1, num))
	}

	// Lakukan broadcast
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
