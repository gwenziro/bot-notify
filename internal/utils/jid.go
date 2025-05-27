package utils

import "strings"

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
