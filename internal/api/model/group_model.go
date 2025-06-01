package model

// GroupParticipantInfo adalah model untuk informasi anggota grup
type GroupParticipantInfo struct {
	JID          string `json:"jid"`                    // JID WhatsApp lengkap
	PhoneNumber  string `json:"phoneNumber,omitempty"`  // Nomor telepon terformat
	IsAdmin      bool   `json:"isAdmin"`                // Flag status admin
	IsSuperAdmin bool   `json:"isSuperAdmin,omitempty"` // Flag status super admin
	ContactName  string `json:"contactName,omitempty"`  // Nama dari kontak
}

// GroupInfo adalah model untuk informasi dasar grup WhatsApp
type GroupInfo struct {
	ID           string                 `json:"id"`           // ID grup WhatsApp
	Name         string                 `json:"name"`         // Nama grup
	MemberCount  int                    `json:"memberCount"`  // Jumlah anggota
	IsAdmin      bool                   `json:"isAdmin"`      // Flag status admin kita
	Participants []GroupParticipantInfo `json:"participants"` // Daftar anggota grup
}

// GroupListResponse adalah model untuk hasil query daftar grup
type GroupListResponse struct {
	BaseResponse
	Count  int         `json:"count"`  // Jumlah grup
	Groups []GroupInfo `json:"groups"` // Daftar grup
}

// NewGroupListResponse membuat instance baru GroupListResponse
func NewGroupListResponse(success bool, message string, groups []GroupInfo) GroupListResponse {
	return GroupListResponse{
		BaseResponse: NewBaseResponse(success, message),
		Count:        len(groups),
		Groups:       groups,
	}
}

// GroupParticipantsResponse adalah model untuk hasil query daftar anggota grup
type GroupParticipantsResponse struct {
	BaseResponse
	GroupID          string                 `json:"groupId"`          // ID grup WhatsApp
	GroupName        string                 `json:"groupName"`        // Nama grup
	ParticipantCount int                    `json:"participantCount"` // Jumlah anggota
	IsAdmin          bool                   `json:"isAdmin"`          // Flag status admin kita
	Participants     []GroupParticipantInfo `json:"participants"`     // Daftar anggota grup
}

// NewGroupParticipantsResponse membuat instance baru GroupParticipantsResponse
func NewGroupParticipantsResponse(message string, groupID string, groupName string, isAdmin bool, participants []GroupParticipantInfo) GroupParticipantsResponse {
	return GroupParticipantsResponse{
		BaseResponse:     NewBaseResponse(true, message),
		GroupID:          groupID,
		GroupName:        groupName,
		ParticipantCount: len(participants),
		IsAdmin:          isAdmin,
		Participants:     participants,
	}
}
