package entity

// LoginPageData merepresentasikan data yang digunakan pada halaman login
type LoginPageData struct {
	Title       string // Judul halaman
	RedirectTo  string // URL redirect setelah login
	Error       string // Pesan error jika ada
	CsrfToken   string // Token CSRF untuk keamanan form
	AccessToken string // Token akses yang dimaskir
}

// LoginRequest merepresentasikan data yang diterima dari form login
type LoginRequest struct {
	Token      string // Token yang dimasukkan pengguna
	RememberMe bool   // Flag untuk mengingat login
	Redirect   string // URL redirect setelah login
	CsrfToken  string // Token CSRF untuk validasi
}

// AutoLoginCookie merepresentasikan konfigurasi cookie auto-login
type AutoLoginCookie struct {
	Name     string
	Value    string
	Path     string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite string
}

// NewLoginPageData membuat instance baru LoginPageData dengan nilai default
func NewLoginPageData(title, redirectTo, errorMsg, csrfToken, accessToken string) LoginPageData {
	if title == "" {
		title = "Login - WhatsApp Bot Notify"
	}

	if redirectTo == "" {
		redirectTo = "/dashboard"
	}

	return LoginPageData{
		Title:       title,
		RedirectTo:  redirectTo,
		Error:       errorMsg,
		CsrfToken:   csrfToken,
		AccessToken: accessToken,
	}
}

// NewLoginRequest membuat instance baru LoginRequest
func NewLoginRequest(token, redirect, csrfToken string, rememberMe bool) LoginRequest {
	return LoginRequest{
		Token:      token,
		RememberMe: rememberMe,
		Redirect:   redirect,
		CsrfToken:  csrfToken,
	}
}

// NewAutoLoginCookie membuat instance baru AutoLoginCookie dengan nilai default
func NewAutoLoginCookie(name, value, path string, maxAge int, secure, httpOnly bool, sameSite string) AutoLoginCookie {
	if path == "" {
		path = "/"
	}

	if sameSite == "" {
		sameSite = "Strict"
	}

	return AutoLoginCookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   secure,
		HTTPOnly: httpOnly,
		SameSite: sameSite,
	}
}
