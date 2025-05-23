package client

import (
	"strings"

	"go.mau.fi/whatsmeow/types"
)

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

// ParseJID mengkonversi string ID menjadi JID WhatsApp
func ParseJID(id string) (types.JID, error) {
	return types.ParseJID(id)
}

// ParsePhoneNumber mengkonversi nomor telepon menjadi JID personal
func ParsePhoneNumber(phoneNumber string) types.JID {
	number := FormatPhoneNumber(phoneNumber)
	return types.NewJID(number, types.DefaultUserServer)
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

// ParseGroupID mengkonversi ID grup menjadi JID grup
func ParseGroupID(groupID string) types.JID {
	id := FormatGroupID(groupID)
	// Hapus @g.us jika ada untuk memastikan format yang benar
	id = strings.TrimSuffix(id, "@g.us")
	return types.NewJID(id, types.GroupServer)
}

// IsValidPersonalJID memeriksa apakah JID adalah JID personal yang valid
func IsValidPersonalJID(jid types.JID) bool {
	return jid.Server == types.DefaultUserServer && jid.User != ""
}

// IsValidGroupJID memeriksa apakah JID adalah JID grup yang valid
func IsValidGroupJID(jid types.JID) bool {
	return jid.Server == types.GroupServer && jid.User != ""
}
