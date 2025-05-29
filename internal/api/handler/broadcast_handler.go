package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/constants"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// BroadcastHandler menangani endpoint broadcast pesan API
type BroadcastHandler struct {
	BaseHandler
}

// NewBroadcastHandler membuat instance baru BroadcastHandler
func NewBroadcastHandler(whatsClient *client.Client) *BroadcastHandler {
	return &BroadcastHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-broadcast"),
	}
}

// SendBroadcast mengirim pesan ke banyak nomor/grup sekaligus
// Endpoint: POST /api/send/broadcast
func (h *BroadcastHandler) SendBroadcast(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "SendBroadcast")

	// 2. Validasi koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgBroadcastConnectionError) {
		return nil
	}

	// 3. Parse dan validasi input
	req, err := h.parseBroadcastRequest(c)
	if err != nil {
		return err
	}

	// 4. Proses target penerima
	cleanedReq, err := h.processBroadcastTargets(c, req)
	if err != nil {
		return err
	}

	// 5. Kirim broadcast dan buat respons
	return h.executeBroadcast(c, cleanedReq)
}

// parseBroadcastRequest mem-parsing dan memvalidasi request broadcast
func (h *BroadcastHandler) parseBroadcastRequest(c *fiber.Ctx) (model.BroadcastRequest, error) {
	var req model.BroadcastRequest

	// Parse request body
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return req, err
	}

	// Validasi pesan
	if req.Message == "" {
		return req, h.SendError(c, constants.MsgEmptyMessage, nil, fiber.StatusBadRequest)
	}

	// Debug log parsed fields
	h.Logger.Debug("Request broadcast diparsing", utils.Fields{
		"personal_count": len(req.PersonalNumbers),
		"group_count":    len(req.GroupIDs),
		"message_len":    len(req.Message),
	})

	return req, nil
}

// processBroadcastTargets memproses dan memvalidasi target broadcast
func (h *BroadcastHandler) processBroadcastTargets(c *fiber.Ctx, req model.BroadcastRequest) (model.BroadcastRequest, error) {
	// Fungsi logger untuk utils
	logFunc := func(msg string, fields utils.Fields) {
		h.Logger.WithFields(fields).Info(msg)
	}

	// Proses dan deduplikasi personalNumbers
	cleanedPersonalNumbers, dupPersonalCount := utils.CleanAndDedupPersonalNumbers(req.PersonalNumbers, logFunc)

	// Proses dan deduplikasi groupIDs
	cleanedGroupIDs, dupGroupCount := utils.CleanAndDedupGroupIDs(req.GroupIDs, logFunc)

	// Log hasil deduplikasi
	if dupPersonalCount > 0 || dupGroupCount > 0 {
		h.Logger.Info("Ditemukan dan dilewati target duplikat", utils.Fields{
			"duplicate_personal": dupPersonalCount,
			"duplicate_groups":   dupGroupCount,
		})
	}

	// Update request dengan data yang sudah dibersihkan
	cleanedReq := req
	cleanedReq.PersonalNumbers = cleanedPersonalNumbers
	cleanedReq.GroupIDs = cleanedGroupIDs

	// Hitung ulang total target
	totalTargets := len(cleanedReq.PersonalNumbers) + len(cleanedReq.GroupIDs)

	// Validasi jumlah minimum target
	if totalTargets < 2 {
		return cleanedReq, h.SendError(c, constants.MsgMinTarget, nil, fiber.StatusBadRequest)
	}

	// Validasi jumlah maksimum target
	if totalTargets > 16 {
		return cleanedReq, h.SendError(c, constants.MsgMaxTarget, nil, fiber.StatusBadRequest)
	}

	// Sesuaikan delay jika tidak dalam rentang yang wajar
	if cleanedReq.DelayMs <= 0 {
		cleanedReq.DelayMs = constants.DefaultBroadcastDelay
	} else if cleanedReq.DelayMs > constants.MaxBroadcastDelay {
		cleanedReq.DelayMs = constants.MaxBroadcastDelay
	}

	return cleanedReq, nil
}

// executeBroadcast menjalankan operasi broadcast dan mengirim respons
func (h *BroadcastHandler) executeBroadcast(c *fiber.Ctx, req model.BroadcastRequest) error {
	// Catat waktu mulai untuk menghitung durasi proses
	startTime := time.Now()

	// Log awal broadcast
	h.Logger.WithFields(utils.Fields{
		"personal_count": len(req.PersonalNumbers),
		"group_count":    len(req.GroupIDs),
		"delay_ms":       req.DelayMs,
	}).Info("Memulai pengiriman broadcast")

	// Lakukan broadcast dengan data yang sudah bersih dan dapatkan waktu pengiriman terakhir
	results, lastSentTime := h.WhatsApp.BroadcastMessage(
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
	h.LogSuccessResponse("Broadcast selesai", utils.Fields{
		"total":           len(results),
		"success":         successCount,
		"failed":          len(results) - successCount,
		"processing_time": processingTime,
	})

	// Buat respons
	response := model.NewBroadcastResponse(
		constants.MsgBroadcastSuccess,
		results,
		processingTime,
		lastSentTime,
	)

	return h.SendSuccess(c, response)
}
