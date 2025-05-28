package model

// ProfileInfo berisi informasi dasar tentang akun WhatsApp yang terhubung
type ProfileInfo struct {
	ID             string `json:"id,omitempty"`
	PhoneNumber    string `json:"phoneNumber,omitempty"`
	Name           string `json:"name,omitempty"`
	Status         string `json:"status,omitempty"`
	IsConnected    bool   `json:"isConnected"`
	IsLoggedIn     bool   `json:"isLoggedIn"`
	PictureURL     string `json:"pictureUrl,omitempty"`
	ConnectedSince string `json:"connectedSince,omitempty"`
}

// ProfileResponse untuk hasil query profil WhatsApp
type ProfileResponse struct {
	BaseResponse
	Profile ProfileInfo `json:"profile"`
}

// NewProfileResponse membuat response profil baru
func NewProfileResponse(success bool, message string, profile ProfileInfo) ProfileResponse {
	return ProfileResponse{
		BaseResponse: NewBaseResponse(success, message),
		Profile:      profile,
	}
}
