package model

// ConnectionStatus adalah model untuk informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status       string `json:"status"`                 // Status koneksi sebagai string
	IsConnected  bool   `json:"isConnected"`            // Flag status koneksi
	LastActivity string `json:"lastActivity,omitempty"` // Waktu aktivitas terakhir
}

// StatusResponse adalah model untuk respons dari endpoint status
type StatusResponse struct {
	BaseResponse
	Status       string `json:"status"`                 // Status koneksi sebagai string
	IsConnected  bool   `json:"isConnected"`            // Flag status koneksi
	LastActivity string `json:"lastActivity,omitempty"` // Waktu aktivitas terakhir
}

// NewStatusResponse membuat instance baru StatusResponse
func NewStatusResponse(message string, status string, isConnected bool) StatusResponse {
	return StatusResponse{
		BaseResponse: NewBaseResponse(isConnected, message),
		Status:       status,
		IsConnected:  isConnected,
	}
}

// PingResponse adalah model untuk respons dari endpoint ping
type PingResponse struct {
	BaseResponse
	Version string `json:"version"` // Versi API
}

// NewPingResponse membuat instance baru PingResponse
func NewPingResponse(message string, version string) PingResponse {
	return PingResponse{
		BaseResponse: NewBaseResponse(true, message),
		Version:      version,
	}
}
