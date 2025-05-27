package model

import (
	"time"
)

// ConnectionStatus berisi informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status            string `json:"status"`
	IsConnected       bool   `json:"isConnected"`                 // Gunakan camelCase
	ConnectionRetries int    `json:"connectionRetries,omitempty"` // Gunakan camelCase
	LastActivity      string `json:"lastActivity,omitempty"`      // Ubah ke string terformat
	Timestamp         string `json:"timestamp,omitempty"`         // Ubah ke string terformat
}

// StatusResponse berisi respons dari endpoint status
type StatusResponse struct {
	Success    bool             `json:"success"`
	Details    ConnectionStatus `json:"details"`
	Time       string           `json:"time"`       // Format waktu yang user-friendly
	ServerTime time.Time        `json:"serverTime"` // Waktu asli untuk perhitungan
}

// PingResponse berisi respons dari endpoint ping
type PingResponse struct {
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	Time          time.Time `json:"time"`
	Version       string    `json:"version"`
	TimeFormatted string    `json:"time_formatted"`
}
