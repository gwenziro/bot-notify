package model

// ConnectionStatus berisi informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status            string `json:"status"`
	IsConnected       bool   `json:"isConnected"`
	ConnectionRetries int    `json:"connectionRetries,omitempty"`
	LastActivity      string `json:"lastActivity,omitempty"`
}

// StatusResponse berisi respons dari endpoint status
type StatusResponse struct {
	BaseResponse
	Status            string `json:"status"`
	IsConnected       bool   `json:"isConnected"`
	ConnectionRetries int    `json:"connectionRetries,omitempty"`
	LastActivity      string `json:"lastActivity,omitempty"`
}

// NewStatusResponse membuat respons status baru
func NewStatusResponse(message string, status string, isConnected bool) StatusResponse {
	resp := StatusResponse{
		BaseResponse: NewBaseResponse(isConnected, message),
		Status:       status,
		IsConnected:  isConnected,
	}

	return resp
}

// PingResponse berisi respons dari endpoint ping
type PingResponse struct {
	BaseResponse
	Version string `json:"version"`
}

// NewPingResponse membuat respons ping baru
func NewPingResponse(message string, version string) PingResponse {
	return PingResponse{
		BaseResponse: NewBaseResponse(true, message),
		Version:      version,
	}
}
