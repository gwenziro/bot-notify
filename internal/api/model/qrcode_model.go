package model

// QRCodeStatusResponse berisi informasi status QR code
type QRCodeStatusResponse struct {
	BaseResponse
	Available       bool `json:"available"`
	Expired         bool `json:"expired"`
	ConnectedStatus bool `json:"connected,omitempty"`
}

// NewQRCodeStatusResponse membuat respons status QR code baru
func NewQRCodeStatusResponse(available bool, expired bool) QRCodeStatusResponse {
	// Tentukan pesan dan status sukses berdasarkan status
	var message string
	success := available && !expired // Success=true hanya jika tersedia dan tidak expired

	if expired {
		message = "QR code sudah kedaluwarsa. Silakan gunakan endpoint /api/reconnect untuk mendapatkan QR code baru"
		success = false // QR expired = tidak sukses
	} else if !available {
		message = "QR code belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu"
		success = false // QR tidak tersedia = tidak sukses
	} else {
		message = "QR code tersedia, silakan pindai"
	}

	return QRCodeStatusResponse{
		BaseResponse: NewBaseResponse(success, message),
		Available:    available,
		Expired:      expired,
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
