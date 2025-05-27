package client

import (
	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// ParseJID mengkonversi string ID menjadi JID WhatsApp
func ParseJID(id string) (types.JID, error) {
	return types.ParseJID(id)
}

// ParsePhoneNumber mengkonversi nomor telepon menjadi JID personal
func ParsePhoneNumber(phoneNumber string) types.JID {
	number := utils.FormatPhoneNumber(phoneNumber)
	return types.NewJID(number, types.DefaultUserServer)
}

// ParseGroupID mengkonversi ID grup menjadi JID grup
func ParseGroupID(groupID string) types.JID {
	id := utils.FormatGroupID(groupID)
	// Hapus @g.us jika ada untuk memastikan format yang benar
	id = id[:len(id)-5] // Menghapus "@g.us"
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

// Alias untuk fungsi di utils package untuk kemudahan penggunaan
var (
	FormatPhoneNumber    = utils.FormatPhoneNumber
	FormatGroupID        = utils.FormatGroupID
	FormatWhatsAppNumber = utils.FormatWhatsAppNumber
	NormalizeJID         = utils.NormalizeJID
)
