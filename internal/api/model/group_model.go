package model

// GroupParticipantInfo berisi informasi tentang anggota grup
type GroupParticipantInfo struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	PushName     string `json:"pushName,omitempty"`
	ContactName  string `json:"contactName,omitempty"`
}

// GroupInfo berisi informasi dasar tentang grup WhatsApp
type GroupInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	MemberCount  int                    `json:"memberCount"`
	IsAdmin      bool                   `json:"isAdmin"`
	Participants []GroupParticipantInfo `json:"participants"`
}

// GroupListResponse untuk hasil query daftar grup
type GroupListResponse struct {
	BaseResponse
	Count  int         `json:"count"`
	Groups []GroupInfo `json:"groups"`
}

// NewGroupListResponse membuat response daftar grup baru
func NewGroupListResponse(message string, groups []GroupInfo) GroupListResponse {
	return GroupListResponse{
		BaseResponse: NewBaseResponse(true, message),
		Count:        len(groups),
		Groups:       groups,
	}
}

// GroupParticipantsResponse untuk hasil query daftar anggota grup
type GroupParticipantsResponse struct {
	BaseResponse
	GroupID          string                 `json:"groupId"`
	GroupName        string                 `json:"groupName"`
	ParticipantCount int                    `json:"participantCount"`
	IsAdmin          bool                   `json:"isAdmin"`
	Participants     []GroupParticipantInfo `json:"participants"`
}

// NewGroupParticipantsResponse membuat response daftar anggota grup baru
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
