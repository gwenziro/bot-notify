package model

import "time"

// ReconnectRequest untuk request menghubungkan kembali WhatsApp
type ReconnectRequest struct {
	Force bool `json:"force"` // Opsional: force reconnect meskipun sudah terhubung
}

// ConnectionResponse adalah respons dasar untuk operasi koneksi
type ConnectionResponse struct {
	Success   bool      `json:"sukses"`
	Message   string    `json:"pesan"`
	Timestamp time.Time `json:"waktu"`
	Status    string    `json:"status,omitempty"`
}

// NewConnectionResponse membuat respons koneksi baru
func NewConnectionResponse(success bool, message string, status string) ConnectionResponse {
	return ConnectionResponse{
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
		Status:    status,
	}
}
