package model

// GroupInfo berisi informasi dasar tentang grup WhatsApp
type GroupInfo struct {
	ID          string `json:"id"`
	Name        string `json:"nama"`
	MemberCount int    `json:"jumlahAnggota,omitempty"`
	IsAdmin     bool   `json:"admin,omitempty"`
}

// GroupListResponse untuk hasil query daftar grup
type GroupListResponse struct {
	Success bool        `json:"sukses"`
	Message string      `json:"pesan"`
	Count   int         `json:"jumlah"`
	Groups  []GroupInfo `json:"grup"`
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
