package model

// ProfileInfo adalah model untuk informasi dasar akun WhatsApp yang terhubung
type ProfileInfo struct {
	ID             string `json:"id,omitempty"`             // ID WhatsApp unik
	PhoneNumber    string `json:"phoneNumber,omitempty"`    // Nomor telepon terformat
	Name           string `json:"name,omitempty"`           // Nama profil/kontak
	Status         string `json:"status,omitempty"`         // Status profil WhatsApp
	IsConnected    bool   `json:"isConnected"`              // Flag status koneksi
	IsLoggedIn     bool   `json:"isLoggedIn"`               // Flag status login
	PictureURL     string `json:"pictureUrl,omitempty"`     // URL foto profil
	ConnectedSince string `json:"connectedSince,omitempty"` // Waktu terhubung dalam format Indonesia
}

// ProfileResponse adalah model untuk hasil query profil WhatsApp
type ProfileResponse struct {
	BaseResponse
	Profile ProfileInfo `json:"profile"` // Informasi profil WhatsApp
}

// NewProfileResponse membuat instance baru ProfileResponse
func NewProfileResponse(success bool, message string, profile ProfileInfo) ProfileResponse {
	return ProfileResponse{
		BaseResponse: NewBaseResponse(success, message),
		Profile:      profile,
	}
}
