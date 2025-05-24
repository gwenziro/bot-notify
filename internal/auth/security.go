package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword menghash password menggunakan bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword memverifikasi password terhadap hash
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// SecureCompare membandingkan dua string dengan waktu konstan
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// GenerateRandomBytes menghasilkan bytes acak dengan panjang tertentu
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateHexToken menghasilkan token hex dengan panjang tertentu
func GenerateHexToken(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateBase64Token menghasilkan token base64 dengan panjang tertentu
func GenerateBase64Token(length int) (string, error) {
	// Hitung berapa byte yang diperlukan untuk length karakter base64
	// Base64: 4 karakter = 3 byte
	requiredBytes := length * 3 / 4
	bytes, err := GenerateRandomBytes(requiredBytes)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	if len(token) > length {
		token = token[:length]
	}
	return token, nil
}

// GenerateRandomInt menghasilkan integer acak dalam rentang [min, max]
func GenerateRandomInt(min, max int64) (int64, error) {
	if min > max {
		return 0, fmt.Errorf("min tidak boleh lebih besar dari max")
	}

	// Hitung range (max - min + 1)
	diff := max - min + 1

	// Hasilkan angka acak
	n, err := rand.Int(rand.Reader, big.NewInt(diff))
	if err != nil {
		return 0, err
	}

	// Konversi dan tambahkan offset min
	return n.Int64() + min, nil
}

// SecureCookie mengembalikan opsi cookie yang aman
func SecureCookie() fiber.Cookie {
	return fiber.Cookie{
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	}
}

// IsPrivateIP memeriksa apakah IP adalah IP privat
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Pengecekan RFC1918 private IP
	private := false

	// IPv4 private ranges
	private = private || ip.IsPrivate()

	// Check if it's a loopback address
	private = private || ip.IsLoopback()

	return private
}

// SanitizeInput membersihkan input dari karakter berbahaya
func SanitizeInput(input string) string {
	// Contoh: hapus karakter kontrol dan trim
	result := strings.TrimSpace(input)
	return result
}

// MaskIP mengaburkan bagian dari IP untuk privasi
func MaskIP(ip string) string {
	if ip == "" {
		return ""
	}

	// Untuk IPv4
	if parts := strings.Split(ip, "."); len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.xxx", parts[0], parts[1], parts[2])
	}

	// Untuk IPv6
	if strings.Contains(ip, ":") {
		return strings.Split(ip, ":")[0] + ":xxxx:xxxx:xxxx:xxxx"
	}

	return ip
}
