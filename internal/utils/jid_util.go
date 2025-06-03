package utils

import (
	"strings"

	"go.mau.fi/whatsmeow/types" // Import untuk types.JID
)

// NormalizeJID menormalkan JID WhatsApp untuk perbandingan yang konsisten
func NormalizeJID(jid string) string {
	// Hapus bagian device identifier jika ada
	normalized := jid

	if idx := strings.IndexByte(normalized, '@'); idx > 0 {
		// Ekstrak bagian nomor
		userPart := normalized[:idx]
		// Ambil bagian server
		serverPart := normalized[idx:]

		// Hapus device identifier
		if deviceIdx := strings.IndexByte(userPart, ':'); deviceIdx > 0 {
			userPart = userPart[:deviceIdx]
		}

		// Gabungkan kembali
		normalized = userPart + serverPart
	}

	return normalized
}

// FormatPhoneNumber memformat nomor telepon menjadi format ID WhatsApp personal
func FormatPhoneNumber(number string) string {
	// Bersihkan nomor dari karakter non-digit
	number = strings.TrimSpace(number)

	// Hapus karakter non-digit (selain + di awal)
	if strings.HasPrefix(number, "+") {
		number = "+" + strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, number[1:])
	} else {
		number = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, number)
	}

	// Jika dimulai dengan +, hapus +
	number = strings.TrimPrefix(number, "+")

	// Jika dimulai dengan 0, ganti dengan 62 (kode negara Indonesia)
	if strings.HasPrefix(number, "0") {
		number = "62" + number[1:]
	}

	return number
}

// FormatGroupID memformat ID grup WhatsApp
func FormatGroupID(id string) string {
	// Jika sudah memiliki @g.us, gunakan apa adanya
	if strings.Contains(id, "@g.us") {
		return id
	}

	// Hapus karakter non-digit dari ID
	id = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, id)

	return id + "@g.us"
}

// FormatWhatsAppNumber mengubah JID WhatsApp menjadi nomor telepon lokal yang mudah dibaca
func FormatWhatsAppNumber(jid string) string {
	// Menangani format XXXXXX@s.whatsapp.net atau XXXXXX:XX@s.whatsapp.net
	parts := strings.Split(jid, "@")
	if len(parts) < 2 {
		return jid
	}

	// Ambil bagian nomor telepon (yang mungkin memiliki device ID setelah :)
	phoneAndDevice := strings.Split(parts[0], ":")
	phone := phoneAndDevice[0]

	// Ubah awalan 62 (Indonesia) menjadi 0
	if strings.HasPrefix(phone, "62") {
		return "0" + phone[2:]
	}

	return phone
}

// IsPersonalJID memeriksa apakah string JID adalah JID personal (bukan grup)
func IsPersonalJID(jid string) bool {
	return strings.Contains(jid, "@s.whatsapp.net")
}

// IsGroupJID memeriksa apakah string JID adalah JID grup
func IsGroupJID(jid string) bool {
	return strings.Contains(jid, "@g.us")
}

// ValidatePhoneNumber memeriksa apakah nomor telepon valid untuk WhatsApp
// Format valid: 628xxxxxxxxxx (kode negara + nomor tanpa awalan 0)
func ValidatePhoneNumber(number string) bool {
	number = FormatPhoneNumber(number)

	// Minimal 10 digit (kode negara + nomor)
	// Maksimal 15 digit (standar ITU-T E.164 untuk nomor internasional)
	if len(number) < 10 || len(number) > 15 {
		return false
	}

	// Harus angka semua
	for _, c := range number {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// ValidateGroupID memeriksa apakah ID grup valid untuk WhatsApp
func ValidateGroupID(id string) bool {
	// Bersihkan ID
	id = strings.TrimSpace(id)

	// Jika kosong, tidak valid
	if id == "" {
		return false
	}

	// Cek jika ID memiliki format array "[123, 456]"
	// dan ekstrak nilai pertama dari array
	if strings.HasPrefix(id, "[") && strings.Contains(id, "]") {
		// Ekstrak konten di dalam tanda kurung
		content := id[1:strings.Index(id, "]")]

		// Split berdasarkan koma
		parts := strings.Split(content, ",")
		if len(parts) > 0 {
			// Gunakan elemen pertama sebagai ID
			id = strings.TrimSpace(parts[0])
		}
	}

	// Jika sudah mengandung @g.us, cek formatnya
	if strings.Contains(id, "@g.us") {
		parts := strings.Split(id, "@")
		return len(parts) == 2 && parts[0] != "" && parts[1] == "g.us"
	}

	// Jika tidak mengandung @g.us, harus angka semua
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}

	return len(id) > 5 // Minimal panjang ID grup
}

// NormalizeGroupID menormalisasi ID grup untuk perbandingan yang konsisten
func NormalizeGroupID(id string) string {
	// Hapus @g.us jika ada
	id = strings.Split(id, "@")[0]

	// Hapus semua karakter non-digit
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, id)
}

// ParseGroupID mengkonversi ID grup menjadi JID grup
func ParseGroupID(groupID string) types.JID {
	// Validasi ID grup tidak boleh kosong
	if !ValidateGroupID(groupID) {
		// Log warning dan gunakan ID placeholder untuk menghindari panic
		Warn("Group ID tidak valid", Fields{"id": groupID})
		return types.NewJID("invalid", types.GroupServer)
	}

	// Jika sudah memiliki @g.us, ekstrak ID-nya saja
	if strings.Contains(groupID, "@g.us") {
		id := strings.Split(groupID, "@")[0]
		return types.NewJID(id, types.GroupServer)
	}

	// Gunakan FormatGroupID untuk mendapatkan ID yang benar
	formattedID := FormatGroupID(groupID)
	id := strings.Split(formattedID, "@")[0] // Hapus bagian @g.us

	return types.NewJID(id, types.GroupServer)
}
