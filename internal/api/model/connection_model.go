package model

// ReconnectRequest adalah model untuk request menghubungkan kembali WhatsApp
type ReconnectRequest struct {
	Force bool `json:"force"` // Flag untuk memaksa koneksi ulang meskipun sudah terhubung
}

// ConnectionResponse adalah model untuk respons operasi koneksi
type ConnectionResponse struct {
	BaseResponse
	Status string `json:"status,omitempty"` // Status koneksi saat ini
}

// NewConnectionResponse membuat instance baru ConnectionResponse
func NewConnectionResponse(success bool, message string, status string) ConnectionResponse {
	return ConnectionResponse{
		BaseResponse: NewBaseResponse(success, message),
		Status:       status,
	}
}
