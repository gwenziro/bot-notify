package model

import "time"

// ConnectionStatus berisi informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status            string    `json:"status"`
	IsConnected       bool      `json:"isConnected"`
	ConnectionRetries int       `json:"connectionRetries"`
	LastActivity      time.Time `json:"lastActivity"`
	Timestamp         time.Time `json:"timestamp"`
}

// StatusResponse untuk hasil query status
type StatusResponse struct {
	Success bool             `json:"sukses"`
	Status  string           `json:"status"`
	Details ConnectionStatus `json:"details"`
	Time    time.Time        `json:"timestamp"`
}

// PingResponse untuk endpoint health check
type PingResponse struct {
	Success bool      `json:"sukses"`
	Message string    `json:"pesan"`
	Time    time.Time `json:"waktu"`
	Version string    `json:"versi,omitempty"`
}
