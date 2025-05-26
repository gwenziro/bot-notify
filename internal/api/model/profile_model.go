package model

import "time"

// ProfileInfo berisi informasi dasar tentang akun WhatsApp yang terhubung
type ProfileInfo struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name"`
	Status      string `json:"status,omitempty"`
	IsConnected bool   `json:"isConnected"`
	IsLoggedIn  bool   `json:"isLoggedIn"`
	PictureURL  string `json:"pictureUrl,omitempty"`
	Device      string `json:"device,omitempty"`
}

// ProfileResponse untuk hasil query profil WhatsApp
type ProfileResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Profile   ProfileInfo `json:"profile"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewProfileResponse membuat response profil baru
func NewProfileResponse(message string, profile ProfileInfo) ProfileResponse {
	return ProfileResponse{
		Success:   true,
		Message:   message,
		Profile:   profile,
		Timestamp: time.Now(),
	}
}

// ProfileErrorResponse untuk response error profil
func NewProfileErrorResponse(message string) ProfileResponse {
	return ProfileResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}
}
