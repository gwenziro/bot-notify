package client

import (
	"strings"

	"github.com/gwenziro/bot-notify/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

// ParseJID mengkonversi string ID menjadi JID WhatsApp
func ParseJID(id string) (types.JID, error) {
	return types.ParseJID(id)
}

// ParsePhoneNumber mengkonversi nomor telepon menjadi JID personal
func ParsePhoneNumber(phoneNumber string) types.JID {
	// Gunakan utils.FormatPhoneNumber untuk format standar
	number := utils.FormatPhoneNumber(phoneNumber)
	return types.NewJID(number, types.DefaultUserServer)
}

// ParseGroupID mengkonversi ID grup menjadi JID grup
func ParseGroupID(groupID string) types.JID {
	// Validasi ID grup tidak boleh kosong menggunakan utils.ValidateGroupID
	if !utils.ValidateGroupID(groupID) {
		// Log warning dan gunakan ID placeholder untuk menghindari panic
		utils.Warn("Group ID tidak valid", utils.Fields{"id": groupID})
		return types.NewJID("invalid", types.GroupServer)
	}

	// Jika sudah memiliki @g.us, ekstrak ID-nya saja
	if strings.Contains(groupID, "@g.us") {
		id := strings.Split(groupID, "@")[0]
		return types.NewJID(id, types.GroupServer)
	}

	// Gunakan utils.FormatGroupID untuk mendapatkan ID yang benar
	groupID = utils.FormatGroupID(groupID)
	id := strings.Split(groupID, "@")[0] // Hapus bagian @g.us

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
	ValidatePhoneNumber  = utils.ValidatePhoneNumber
	ValidateGroupID      = utils.ValidateGroupID
)
