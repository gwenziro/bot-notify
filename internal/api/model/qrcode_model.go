package model

import "time"

// QRCodeStatusResponse berisi informasi status QR code
type QRCodeStatusResponse struct {
	BaseResponse
	Available bool       `json:"available"`
	Expired   bool       `json:"expired"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// NewQRCodeStatusResponse membuat respons status QR code baru
func NewQRCodeStatusResponse(available bool, expired bool, timestamp time.Time) QRCodeStatusResponse {
	var timestampPtr *time.Time

	// Hanya sertakan timestamp jika QR code tersedia dan tidak kosong
	if available && !timestamp.IsZero() {
		timestampPtr = &timestamp
	}

	// Tentukan pesan berdasarkan status
	var message string
	if expired {
		message = "QR code sudah kedaluwarsa. Silakan gunakan endpoint /api/reconnect untuk mendapatkan QR code baru"
	} else if !available {
		message = "QR code belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu"
	} else {
		message = "QR code tersedia, silakan pindai"
	}

	return QRCodeStatusResponse{
		BaseResponse: NewBaseResponse(true, message),
		Available:    available,
		Expired:      expired,
		Timestamp:    timestampPtr,
	}
}

// QRCodeErrorResponse untuk respons error terkait QR code
type QRCodeErrorResponse struct {
	BaseErrorResponse
}

// NewQRCodeErrorResponse membuat respons error QR code baru
func NewQRCodeErrorResponse(message string, err error, code int) QRCodeErrorResponse {
	return QRCodeErrorResponse{
		BaseErrorResponse: NewBaseErrorResponse(message, err, code),
	}
}
