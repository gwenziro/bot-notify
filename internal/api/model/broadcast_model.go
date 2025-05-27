package model

import "time"

// BroadcastRequest untuk request API broadcast pesan
type BroadcastRequest struct {
	PersonalNumbers []string `json:"personalNumbers"` // Daftar nomor personal
	GroupIDs        []string `json:"groupIds"`        // Daftar ID grup
	Message         string   `json:"message"`         // Pesan yang akan dikirim
	DelayMs         int      `json:"delayMs"`         // Delay antar pengiriman (ms)
}

// BroadcastResult menyimpan hasil pengiriman ke satu target
type BroadcastResult struct {
	Target   string `json:"target"`          // Nomor/ID target
	Type     string `json:"type"`            // "personal" atau "group"
	Success  bool   `json:"success"`         // Status keberhasilan
	ErrorMsg string `json:"error,omitempty"` // Pesan error jika gagal
}

// BroadcastResponse untuk hasil operasi broadcast
type BroadcastResponse struct {
	BaseResponse
	TotalTargets     int               `json:"totalTargets"`     // Jumlah total target
	SuccessCount     int               `json:"successCount"`     // Jumlah berhasil
	FailedCount      int               `json:"failedCount"`      // Jumlah gagal
	Results          []BroadcastResult `json:"results"`          // Detail hasil
	ProcessingTimeMs int64             `json:"processingTimeMs"` // Waktu pemrosesan
	SentTime         string            `json:"sentTime"`         // Waktu pengiriman
}

// NewBroadcastResponse membuat response broadcast baru
func NewBroadcastResponse(message string, results []BroadcastResult, processingTime int64) BroadcastResponse {
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
		SentTime:         time.Now().Format("02 Jan 2006 15:04:05"),
	}
}
