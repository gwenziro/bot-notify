package model

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
)

// BroadcastRequest adalah model untuk request API broadcast pesan
type BroadcastRequest struct {
	PersonalNumbers []string `json:"personalNumbers"` // Daftar nomor personal penerima
	GroupIDs        []string `json:"groupIds"`        // Daftar ID grup penerima
	Message         string   `json:"message"`         // Pesan yang akan dikirim
	DelayMs         int      `json:"delayMs"`         // Delay antar pengiriman (ms)
}

// BroadcastResult adalah model untuk hasil pengiriman ke satu target
type BroadcastResult struct {
	Target   string `json:"target"`          // Nomor/ID target penerima
	Type     string `json:"type"`            // Tipe target: "personal" atau "group"
	Success  bool   `json:"success"`         // Status keberhasilan pengiriman
	ErrorMsg string `json:"error,omitempty"` // Pesan error jika gagal
}

// BroadcastResponse adalah model untuk hasil operasi broadcast
type BroadcastResponse struct {
	BaseResponse
	TotalTargets     int               `json:"totalTargets"`     // Jumlah total target penerima
	SuccessCount     int               `json:"successCount"`     // Jumlah pengiriman berhasil
	FailedCount      int               `json:"failedCount"`      // Jumlah pengiriman gagal
	Results          []BroadcastResult `json:"results"`          // Detail hasil per target
	ProcessingTimeMs int64             `json:"processingTimeMs"` // Waktu pemrosesan dalam ms
	SentTime         string            `json:"sentTime"`         // Waktu pengiriman dalam format Indonesia
}

// NewBroadcastResponse membuat instance baru BroadcastResponse
func NewBroadcastResponse(message string, results []BroadcastResult, processingTime int64, sentTime time.Time) BroadcastResponse {
	// Hitung jumlah sukses dan gagal
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	return BroadcastResponse{
		BaseResponse:     NewBaseResponse(true, message),
		TotalTargets:     len(results),
		SuccessCount:     successCount,
		FailedCount:      len(results) - successCount,
		Results:          results,
		ProcessingTimeMs: processingTime,
		SentTime:         utils.FormatTimeIndonesia(&sentTime),
	}
}
