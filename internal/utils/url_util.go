package utils

import (
	"net/url"
	"strings"
)

// CleanBaseURL menghapus port dari URL jika ada dan memastikan trailing slash tidak ada
func CleanBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	// Parse URL untuk memisahkan komponen-komponennya
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// Jika parsing gagal, kembalikan URL asli
		return baseURL
	}

	// Hapus port dari host jika ada
	host := parsedURL.Host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Rekonstruksi URL tanpa port
	parsedURL.Host = host

	// Pastikan tidak ada trailing slash
	result := parsedURL.String()
	return strings.TrimSuffix(result, "/")
}

// FormatEndpointURL menambahkan baseURL ke endpoint dengan penanganan slash yang benar
func FormatEndpointURL(baseURL, endpoint string) string {
	// Bersihkan baseURL terlebih dahulu (hapus port)
	cleanBase := CleanBaseURL(baseURL)

	// Pastikan endpoint dimulai dengan / dan tidak ada trailing slash
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	endpoint = strings.TrimSuffix(endpoint, "/")

	return cleanBase + endpoint
}
