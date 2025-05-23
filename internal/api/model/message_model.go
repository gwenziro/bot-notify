package model

import "time"

// PersonalMessageRequest untuk request API kirim pesan personal
type PersonalMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required"`
	Message     string `json:"message" validate:"required"`
}

// GroupMessageRequest untuk request API kirim pesan grup
type GroupMessageRequest struct {
	GroupID string `json:"groupID" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// MessageResponse untuk hasil operasi kirim pesan
type MessageResponse struct {
	Success   bool      `json:"sukses"`
	Message   string    `json:"pesan"`
	Recipient string    `json:"penerima"`
	Type      string    `json:"tipe"`
	Timestamp time.Time `json:"waktu"`
}

// NewMessageResponse membuat respons pesan baru
func NewMessageResponse(message string, recipient string, messageType string) MessageResponse {
	return MessageResponse{
		Success:   true,
		Message:   message,
		Recipient: recipient,
		Type:      messageType,
		Timestamp: time.Now(),
	}
}

// ErrorMessageResponse untuk respons error
type ErrorMessageResponse struct {
	Success   bool      `json:"sukses"`
	Message   string    `json:"pesan"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"waktu"`
	Code      int       `json:"kode"`
}

// NewErrorMessageResponse membuat respons error
func NewErrorMessageResponse(message string, err error, code int) ErrorMessageResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return ErrorMessageResponse{
		Success:   false,
		Message:   message,
		Error:     errMsg,
		Timestamp: time.Now(),
		Code:      code,
	}
}
