package utils

import (
	"strings"
)

// MaskToken menyembunyikan sebagian token untuk keamanan
// Menampilkan hanya 4 karakter pertama dan 4 karakter terakhir
// Contoh: "abcdefghijklmnop" -> "abcd...mnop"
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}

	// Tampilkan hanya 4 karakter pertama dan 4 terakhir
	return token[:4] + "..." + token[len(token)-4:]
}

// TruncateString memotong string jika melebihi panjang maksimum dan menambahkan ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// JoinNonEmpty menggabungkan string non-kosong dengan pemisah yang diberikan
func JoinNonEmpty(separator string, parts ...string) string {
	var nonEmpty []string
	for _, part := range parts {
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, separator)
}
