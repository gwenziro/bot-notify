package model

// QRCodeStatusResponse berisi informasi status QR code
type QRCodeStatusResponse struct {
	BaseResponse
	Available       bool `json:"available"`
	Expired         bool `json:"expired"`
	ConnectedStatus bool `json:"connected,omitempty"`
}

// NewQRCodeStatusResponse membuat respons status QR code baru
func NewQRCodeStatusResponse(available bool, expired bool, message string) QRCodeStatusResponse {
	// Tentukan status sukses berdasarkan status
	success := available && !expired // Success=true hanya jika tersedia dan tidak expired

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
