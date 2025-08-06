package model

import (
	"time"

	"github.com/gwenziro/bot-notify/internal/utils"
)

// BaseResponse adalah model dasar untuk semua respons API
type BaseResponse struct {
	Success bool   `json:"success"` // Status keberhasilan operasi
	Message string `json:"message"` // Pesan status dalam bahasa manusia
	Time    string `json:"time"`    // Waktu respons dalam format Indonesia
}

// NewBaseResponse membuat instance baru BaseResponse dengan nilai default
func NewBaseResponse(success bool, message string) BaseResponse {
	now := time.Now()
	return BaseResponse{
		Success: success,
		Message: message,
		Time:    utils.FormatTimeIndonesia(&now),
	}
}

// BaseErrorResponse adalah model untuk respons error API
type BaseErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"` // Detail error jika ada
	Code  int    `json:"code"`            // HTTP status code
}

// NewBaseErrorResponse membuat instance baru BaseErrorResponse dengan nilai default
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
