package model

import (
	"time"
)

// ConnectionStatus berisi informasi status koneksi WhatsApp
type ConnectionStatus struct {
	Status            string `json:"status"`
	IsConnected       bool   `json:"isConnected"`
	ConnectionRetries int    `json:"connectionRetries,omitempty"`
	LastActivity      string `json:"lastActivity,omitempty"`
	Timestamp         string `json:"timestamp,omitempty"`
}

// StatusResponse berisi respons dari endpoint status
type StatusResponse struct {
	BaseResponse
	Details    ConnectionStatus `json:"details"`
	Time       string           `json:"time"`
	ServerTime time.Time        `json:"serverTime"`
}

// NewStatusResponse membuat respons status baru
func NewStatusResponse(message string, details ConnectionStatus) StatusResponse {
	return StatusResponse{
		BaseResponse: NewBaseResponse(true, message),
		Details:      details,
		Time:         time.Now().Format("02 Jan 2006 15:04:05"),
		ServerTime:   time.Now(),
	}
}

// PingResponse berisi respons dari endpoint ping
type PingResponse struct {
	BaseResponse
	Version       string    `json:"version"`
	Time          time.Time `json:"time"`
	TimeFormatted string    `json:"timeFormatted"`
}

// NewPingResponse membuat respons ping baru
func NewPingResponse(message string, version string) PingResponse {
	now := time.Now()
	return PingResponse{
		BaseResponse:  NewBaseResponse(true, message),
		Version:       version,
		Time:          now,
		TimeFormatted: now.Format("02 Jan 2006 15:04:05"),
	}
}
