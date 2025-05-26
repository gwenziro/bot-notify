package model

import "time"

// QRCodeStatusResponse untuk respons status QR code
type QRCodeStatusResponse struct {
	Success   bool      `json:"success"`
	Available bool      `json:"available"`
	Expired   bool      `json:"expired"`
	Timestamp time.Time `json:"timestamp"`
}

// NewQRCodeStatusResponse membuat respons status QR code baru
func NewQRCodeStatusResponse(available bool, expired bool, timestamp time.Time) QRCodeStatusResponse {
	return QRCodeStatusResponse{
		Success:   true,
		Available: available,
		Expired:   expired,
		Timestamp: timestamp,
	}
}

// QRCodeErrorResponse untuk respons error terkait QR code
type QRCodeErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewQRCodeErrorResponse membuat respons error QR code baru
func NewQRCodeErrorResponse(message string) QRCodeErrorResponse {
	return QRCodeErrorResponse{
		Success: false,
		Message: message,
	}
}
