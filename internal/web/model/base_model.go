package model

import "time"

// BasePageModel berisi field dasar untuk semua halaman web
type BasePageModel struct {
	Title           string
	ActivePage      string
	CurrentYear     int
	IsAuthenticated bool
	IsConnected     bool
	FlashMessage    string
	FlashType       string
}

// NewBasePageModel membuat instance model halaman dasar
func NewBasePageModel(title string, activePage string) BasePageModel {
	return BasePageModel{
		Title:       title,
		ActivePage:  activePage,
		CurrentYear: time.Now().Year(),
	}
}
