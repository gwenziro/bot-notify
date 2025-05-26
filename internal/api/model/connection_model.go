package model

import "time"

// ReconnectRequest untuk request menghubungkan kembali WhatsApp
type ReconnectRequest struct {
	Force bool `json:"force"`
}

// ConnectionResponse adalah respons dasar untuk operasi koneksi
type ConnectionResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
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
