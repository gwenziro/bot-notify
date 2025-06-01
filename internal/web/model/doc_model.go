package model

// EndpointInfo berisi informasi tentang endpoint API
type EndpointInfo struct {
	Name        string // Nama endpoint
	Endpoint    string // Path endpoint
	Method      string // Metode HTTP (GET, POST, dll)
	MethodLower string // Metode HTTP dalam lowercase untuk class CSS
	Description string // Deskripsi endpoint
	Example     string // Contoh penggunaan
	Response    string // Contoh respons
}

// DocumentationModel berisi data yang diperlukan untuk halaman dokumentasi
type DocumentationModel struct {
	BasePageModel
	BaseURL       string         // URL dasar aplikasi
	MaskedToken   string         // Token API yang disamarkan
	Endpoints     []EndpointInfo // Informasi endpoint API
	ActiveSection string         // Bagian dokumentasi yang aktif
}

// NewDocumentationModel membuat instance baru DocumentationModel
func NewDocumentationModel(baseURL string, maskedToken string, endpoints []EndpointInfo) DocumentationModel {
	return DocumentationModel{
		BasePageModel: NewBasePageModel("Dokumentasi API", "docs"),
		BaseURL:       baseURL,
		MaskedToken:   maskedToken,
		Endpoints:     endpoints,
		ActiveSection: "overview", // Default section
	}
}
