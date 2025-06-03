package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// BroadcastHandler menangani endpoint broadcast API
type BroadcastHandler struct {
	BaseHandler
}

// NewBroadcastHandler membuat instance baru BroadcastHandler
func NewBroadcastHandler(whatsClient *client.Client) *BroadcastHandler {
	return &BroadcastHandler{
		BaseHandler: NewBaseHandler(whatsClient, "handler-broadcast"),
	}
}

// Broadcast mengirim pesan ke beberapa nomor/grup sekaligus
// Endpoint: POST /api/broadcast
func (h *BroadcastHandler) SendBroadcast(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "Broadcast")

	// 2. Periksa koneksi WhatsApp
	if !h.CheckWhatsAppConnection(c, constants.MsgBroadcastConnectionError) {
		return nil
	}

	// 3. Parse request
	var req model.BroadcastRequest
	if err := h.ParseAndValidateBody(c, &req); err != nil {
		return err
	}

	// 4. Validasi pesan
	if req.Message == "" {
		return h.SendError(c, constants.MsgEmptyMessage, nil, fiber.StatusBadRequest)
	}

	// 5. Validasi target penerima
	totalTargets := len(req.PersonalNumbers) + len(req.GroupIDs)
	if totalTargets < 2 {
		return h.SendError(c, constants.MsgMinTarget, nil, fiber.StatusBadRequest)
	}

	// 6. Validasi jumlah maksimum target
	if totalTargets > 16 {
		return h.SendError(c, constants.MsgMaxTarget, nil, fiber.StatusBadRequest)
	}

	// 7. Set delay default jika tidak ditentukan
	delayMs := req.DelayMs
	if delayMs <= 0 {
		delayMs = constants.DefaultBroadcastDelay
	} else if delayMs > constants.MaxBroadcastDelay {
		delayMs = constants.MaxBroadcastDelay
	}

	// 8. Mulai operasi broadcast
	startTime := time.Now()

	// Tambahan: validasi format grup ID
	for i, groupID := range req.GroupIDs {
		// Trim spasi
		groupID = strings.TrimSpace(groupID)
		req.GroupIDs[i] = groupID

		// Log untuk debugging
		h.Logger.Debug("Validasi grup ID", utils.Fields{
			"index":    i,
			"group_id": groupID, // Pastikan logging hanya menampilkan ID sebagai string
			"valid":    utils.ValidateGroupID(groupID),
		})
	}

	// 9. Proses broadcast - menerima InternalBroadcastResult dari service
	internalResults, sentTime := h.WhatsApp.BroadcastMessage(
		req.PersonalNumbers,
		req.GroupIDs,
		req.Message,
		delayMs,
	)

	// Transformasi hasil dari service internal ke model API
	apiResults := make([]model.BroadcastResult, len(internalResults))
	for i, result := range internalResults {
		apiResults[i] = model.BroadcastResult{
			Target:   result.Target,
			Type:     result.Type,
			Success:  result.Success,
			ErrorMsg: result.ErrorMsg,
		}

		if result.Success && !result.SentTime.IsZero() {
			apiResults[i].SentTime = utils.FormatTimeIndonesia(&result.SentTime)
		}
	}

	// 10. Hitung waktu eksekusi
	endTime := time.Now()
	processingTime := endTime.Sub(startTime).Milliseconds()

	// 11. Log informasi hasil
	successCount := 0
	for _, result := range internalResults {
		if result.Success {
			successCount++
		}
	}

	h.LogSuccessResponse("Operasi broadcast berhasil", utils.Fields{
		"total":      totalTargets,
		"success":    successCount,
		"failed":     totalTargets - successCount,
		"time_taken": fmt.Sprintf("%dms", processingTime),
	})

	// 12. Kembalikan respons ke client
	return h.SendSuccess(c, model.NewBroadcastResponse(
		constants.MsgBroadcastSuccess,
		apiResults,
		processingTime,
		sentTime,
	))
}
