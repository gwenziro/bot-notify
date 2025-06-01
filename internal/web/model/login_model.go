package model

// LoginModel berisi data yang diperlukan untuk halaman login
type LoginModel struct {
	BasePageModel
	RedirectTo  string // URL tujuan setelah login
	Error       string // Pesan error jika ada
	CsrfToken   string // Token CSRF untuk keamanan form
	AccessToken string // Token API (disamarkan)
}

// NewLoginModel membuat instance baru LoginModel
func NewLoginModel(redirectTo string, error string, csrfToken string) LoginModel {
	return LoginModel{
		BasePageModel: NewBasePageModel("Login", "login"),
		RedirectTo:    redirectTo,
		Error:         error,
		CsrfToken:     csrfToken,
	}
}
