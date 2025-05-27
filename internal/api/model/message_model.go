package model

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
)

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
	BaseResponse
	Recipient string `json:"recipient"`
	Type      string `json:"type"`
	SentTime  string `json:"sentTime,omitempty"`
}

// NewMessageResponse membuat respons pesan baru
func NewMessageResponse(message string, recipient string, messageType string) MessageResponse {
	now := time.Now()

	return MessageResponse{
		BaseResponse: NewBaseResponse(true, message),
		Recipient:    recipient,
		Type:         messageType,
		SentTime:     utils.FormatTimeIndonesia(&now),
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
