package model

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
)

// PersonalMessageRequest adalah model untuk request API kirim pesan personal
type PersonalMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required"` // Nomor telepon penerima
	Message     string `json:"message" validate:"required"`     // Pesan yang akan dikirim
}

// GroupMessageRequest adalah model untuk request API kirim pesan grup
type GroupMessageRequest struct {
	GroupID string `json:"groupID" validate:"required"` // ID grup penerima
	Message string `json:"message" validate:"required"` // Pesan yang akan dikirim
}

// MessageResponse adalah model untuk hasil operasi kirim pesan
type MessageResponse struct {
	BaseResponse
	Recipient string `json:"recipient"`          // ID/nomor penerima pesan
	Type      string `json:"type"`               // Tipe penerima: "personal" atau "group"
	SentTime  string `json:"sentTime,omitempty"` // Waktu pengiriman dalam format Indonesia
}

// NewMessageResponse membuat instance baru MessageResponse
func NewMessageResponse(message string, recipient string, messageType string, sentTime time.Time) MessageResponse {
	return MessageResponse{
		BaseResponse: NewBaseResponse(true, message),
		Recipient:    recipient,
		Type:         messageType,
		SentTime:     utils.FormatTimeIndonesia(&sentTime),
	}
}

// ErrorMessageResponse untuk respons error
type ErrorMessageResponse struct {
	BaseErrorResponse
}

// NewErrorMessageResponse membuat respons error
func NewErrorMessageResponse(message string, err error, code int) ErrorMessageResponse {
	return ErrorMessageResponse{
		BaseErrorResponse: NewBaseErrorResponse(message, err, code),
	}
}
