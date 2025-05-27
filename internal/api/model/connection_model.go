package model

// ReconnectRequest untuk request menghubungkan kembali WhatsApp
type ReconnectRequest struct {
	Force bool `json:"force"`
}

// ConnectionResponse adalah respons dasar untuk operasi koneksi
type ConnectionResponse struct {
	BaseResponse
	Status string `json:"status,omitempty"`
}

// NewConnectionResponse membuat respons koneksi baru
func NewConnectionResponse(success bool, message string, status string) ConnectionResponse {
	return ConnectionResponse{
		BaseResponse: NewBaseResponse(success, message),
		Status:       status,
	}
}
