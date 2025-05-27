package model

import "time"

// BaseResponse adalah model dasar untuk semua respons API
type BaseResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewBaseResponse membuat respons dasar baru
func NewBaseResponse(success bool, message string) BaseResponse {
	return BaseResponse{
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// ErrorResponse untuk respons error
type BaseErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"`
	Code  int    `json:"code"`
}

// NewErrorResponse membuat respons error baru
func NewBaseErrorResponse(message string, err error, code int) BaseErrorResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return BaseErrorResponse{
		BaseResponse: NewBaseResponse(false, message),
		Error:        errMsg,
		Code:         code,
	}
}
