package model

// ErrorModel berisi data yang diperlukan untuk halaman error
type ErrorModel struct {
	BasePageModel
	ErrorCode    int    // Kode HTTP error
	ErrorMessage string // Pesan error
	ReturnPath   string // Path untuk kembali (biasanya "/")
}

// NewErrorModel membuat instance baru ErrorModel
func NewErrorModel(code int, message string) ErrorModel {
	return ErrorModel{
		BasePageModel: NewBasePageModel("Error", "error"),
		ErrorCode:     code,
		ErrorMessage:  message,
		ReturnPath:    "/",
	}
}
