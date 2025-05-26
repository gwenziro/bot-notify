package model

import (
	"time"
)

// ConnectionStatus berisi informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status            string    `json:"status"`
	IsConnected       bool      `json:"is_connected"`
	ConnectionRetries int       `json:"connection_retries"`
	LastActivity      time.Time `json:"last_activity"`
	Timestamp         time.Time `json:"timestamp"`

	// Tambahkan bidang yang diformat untuk waktu
	LastActivityFormatted string `json:"last_activity_formatted"`
	TimestampFormatted    string `json:"timestamp_formatted"`
}

// StatusResponse berisi respons dari endpoint status
type StatusResponse struct {
	Success bool             `json:"success"`
	Status  string           `json:"status"`
	Details ConnectionStatus `json:"details"`
	Time    time.Time        `json:"time"`

	// Tambahkan bidang yang diformat untuk waktu
	TimeFormatted string `json:"time_formatted"`
}

// PingResponse berisi respons dari endpoint ping
type PingResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Version string    `json:"version"`

	// Tambahkan bidang yang diformat untuk waktu
	TimeFormatted string `json:"time_formatted"`
}
