package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken menghasilkan token acak dengan panjang tertentu
func GenerateRandomToken(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// MaskToken menyembunyikan sebagian besar token dengan * untuk keamanan
// Hanya menampilkan beberapa karakter awal dan akhir
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}

	// Tampilkan hanya 4 karakter pertama dan 4 terakhir
	return token[:4] + "..." + token[len(token)-4:]
}
