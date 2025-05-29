package model

// QRCodeStatusResponse adalah model untuk informasi status QR code
type QRCodeStatusResponse struct {
	BaseResponse
	Available       bool `json:"available"`           // Flag ketersediaan QR code
	Expired         bool `json:"expired"`             // Flag kedaluwarsa QR code
	ConnectedStatus bool `json:"connected,omitempty"` // Flag status koneksi WhatsApp
}

// NewQRCodeStatusResponse membuat instance baru QRCodeStatusResponse
func NewQRCodeStatusResponse(available bool, expired bool, message string) QRCodeStatusResponse {
	// Success true hanya jika QR code tersedia dan tidak kedaluwarsa
	success := available && !expired

	return QRCodeStatusResponse{
		BaseResponse: NewBaseResponse(success, message),
		Available:    available,
		Expired:      expired,
	}
}

// QRCodeErrorResponse adalah model untuk respons error terkait QR code
type QRCodeErrorResponse struct {
	BaseErrorResponse
}

// NewQRCodeErrorResponse membuat instance baru QRCodeErrorResponse
func NewQRCodeErrorResponse(message string, err error, code int) QRCodeErrorResponse {
	return QRCodeErrorResponse{
		BaseErrorResponse: NewBaseErrorResponse(message, err, code),
	}
}
