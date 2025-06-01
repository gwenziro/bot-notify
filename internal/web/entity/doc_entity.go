package entity

import (
	"strings"
	"time"
)

// ApiEndpoint merepresentasikan informasi tentang endpoint API
type ApiEndpoint struct {
	Name        string // Nama endpoint
	Endpoint    string // Path endpoint
	Method      string // HTTP method (GET, POST, dll)
	MethodLower string // HTTP method lowercase untuk template
	Description string // Deskripsi endpoint
	Example     string // Contoh penggunaan
	Response    string // Contoh respons
}

// DocumentationData merepresentasikan data untuk halaman dokumentasi
type DocumentationData struct {
	Title             string        // Judul halaman
	CurrentYear       int           // Tahun saat ini untuk footer
	BaseURL           string        // URL dasar aplikasi
	MaskedToken       string        // Token API yang dimaskir
	ActivePage        string        // Halaman aktif untuk navigasi
	WhatsAppConnected bool          // Status koneksi WhatsApp untuk UI
	Endpoints         []ApiEndpoint // Daftar endpoint API
}

// NewDocumentationData membuat instance baru DocumentationData dengan nilai default
func NewDocumentationData(baseURL, maskedToken string, isConnected bool) DocumentationData {
	return DocumentationData{
		Title:             "Dokumentasi API",
		CurrentYear:       time.Now().Year(),
		BaseURL:           baseURL,
		MaskedToken:       maskedToken,
		ActivePage:        "docs",
		WhatsAppConnected: isConnected,
		Endpoints:         []ApiEndpoint{},
	}
}

// NewApiEndpoint membuat instance baru ApiEndpoint
func NewApiEndpoint(name, endpoint, method, description, example, response string) ApiEndpoint {
	return ApiEndpoint{
		Name:        name,
		Endpoint:    endpoint,
		Method:      method,
		MethodLower: strings.ToLower(method),
		Description: description,
		Example:     example,
		Response:    response,
	}
}
