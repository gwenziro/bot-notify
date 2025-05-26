package model

// GroupInfo berisi informasi dasar tentang grup WhatsApp
type GroupInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount,omitempty"`
	IsAdmin     bool   `json:"isAdmin"` // Hapus omitempty agar selalu muncul
}

// GroupListResponse untuk hasil query daftar grup
type GroupListResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Count   int         `json:"count"`
	Groups  []GroupInfo `json:"groups"`
}

// NewGroupListResponse membuat response daftar grup baru
func NewGroupListResponse(message string, groups []GroupInfo) GroupListResponse {
	return GroupListResponse{
		Success: true,
		Message: message,
		Count:   len(groups),
		Groups:  groups,
	}
}
