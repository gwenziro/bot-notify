package utils

import (
	"fmt"
	"regexp"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

// ParsePhoneNumber mengubah nomor telepon menjadi JID WhatsApp yang valid
func ParsePhoneNumber(phone string) types.JID {
	// Bersihkan nomor telepon
	phone = cleanPhoneNumber(phone)

	// Konversi ke JID WhatsApp
	return types.NewJID(phone, types.DefaultUserServer)
}

// ParseGroupID mengubah ID grup menjadi JID grup WhatsApp yang valid
func ParseGroupID(groupID string) types.JID {
	// Bersihkan ID grup
	groupID = cleanGroupID(groupID)

	// Konversi ke JID grup WhatsApp
	return types.NewJID(groupID, types.GroupServer)
}

// cleanPhoneNumber membersihkan dan memformat nomor telepon
func cleanPhoneNumber(phone string) string {
	// Hapus semua karakter non-digit kecuali +
	reg := regexp.MustCompile(`[^\d+]`)
	phone = reg.ReplaceAllString(phone, "")

	// Hapus + di awal jika ada
	phone = strings.TrimPrefix(phone, "+")

	// Jika dimulai dengan 0, ganti dengan 62 (kode negara Indonesia)
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}

	// Jika tidak dimulai dengan kode negara, tambahkan 62
	if !strings.HasPrefix(phone, "62") && len(phone) >= 8 {
		phone = "62" + phone
	}

	return phone
}

// cleanGroupID membersihkan dan memformat ID grup
func cleanGroupID(groupID string) string {
	// Hapus @g.us jika sudah ada
	groupID = strings.Split(groupID, "@")[0]

	// Hapus semua karakter non-digit
	reg := regexp.MustCompile(`\D`)
	groupID = reg.ReplaceAllString(groupID, "")

	return groupID
}

// FormatWhatsAppNumber mengubah JID WhatsApp menjadi nomor telepon yang mudah dibaca
func FormatWhatsAppNumber(jid string) string {
	// Ambil bagian nomor dari JID (sebelum @)
	parts := strings.Split(jid, "@")
	if len(parts) == 0 {
		return jid
	}

	phone := parts[0]

	// Hapus device ID jika ada (setelah :)
	phone = strings.Split(phone, ":")[0]

	// Konversi kode negara Indonesia 62 ke format lokal 0
	if strings.HasPrefix(phone, "62") && len(phone) > 2 {
		return "0" + phone[2:]
	}

	return phone
}

// IsValidPhoneNumber memeriksa apakah nomor telepon valid
func IsValidPhoneNumber(phone string) bool {
	phone = cleanPhoneNumber(phone)

	// Harus memiliki minimal 10 digit (termasuk kode negara)
	if len(phone) < 10 {
		return false
	}

	// Harus dimulai dengan kode negara yang valid
	// Indonesia: 62, Malaysia: 60, Singapura: 65, dll
	validCountryCodes := []string{"62", "60", "65", "66", "84", "95", "98"}
	for _, code := range validCountryCodes {
		if strings.HasPrefix(phone, code) {
			return true
		}
	}

	return false
}

// IsValidGroupID memeriksa apakah ID grup valid
func IsValidGroupID(groupID string) bool {
	groupID = cleanGroupID(groupID)

	// ID grup harus minimal 15 digit
	if len(groupID) < 15 {
		return false
	}

	// Harus berupa angka semua
	reg := regexp.MustCompile(`^\d+$`)
	return reg.MatchString(groupID)
}

// ValidateMessageRecipient memvalidasi penerima pesan (dipindahkan dari utils.go)
func ValidateMessageRecipient(recipient string) (types.JID, string, error) {
	recipient = strings.TrimSpace(recipient)

	if recipient == "" {
		return types.EmptyJID, "", fmt.Errorf("penerima tidak boleh kosong")
	}

	// Cek apakah ini grup atau personal
	if strings.Contains(recipient, "@g.us") || (!strings.Contains(recipient, "@") && len(recipient) > 10) {
		// Ini adalah grup
		if !IsValidGroupID(recipient) {
			return types.EmptyJID, "", fmt.Errorf("ID grup tidak valid")
		}
		jid := ParseGroupID(recipient)
		return jid, "group", nil
	} else {
		// Ini adalah nomor personal
		if !IsValidPhoneNumber(recipient) {
			return types.EmptyJID, "", fmt.Errorf("nomor telepon tidak valid")
		}
		jid := ParsePhoneNumber(recipient)
		return jid, "personal", nil
	}
}

// FormatJIDForDisplay memformat JID untuk ditampilkan ke user (dipindahkan dari utils.go)
func FormatJIDForDisplay(jid types.JID) string {
	if jid.Server == types.GroupServer {
		// Untuk grup, tampilkan ID grup
		return jid.User + "@g.us"
	} else {
		// Untuk personal, format sebagai nomor telepon lokal
		return FormatWhatsAppNumber(jid.String())
	}
}

// GetJIDType menentukan tipe JID (dipindahkan dari utils.go)
func GetJIDType(jid types.JID) string {
	if jid.Server == types.GroupServer {
		return "group"
	}
	return "personal"
}

// NormalizeJID menormalkan JID untuk perbandingan (mengupdate implementasi yang sudah ada)
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
