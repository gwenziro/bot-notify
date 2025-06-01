package model

// HomeModel berisi data yang diperlukan untuk halaman beranda
type HomeModel struct {
	BasePageModel
	Description string // Deskripsi aplikasi
	Version     string // Versi aplikasi
}

// NewHomeModel membuat instance baru HomeModel
func NewHomeModel(version string) HomeModel {
	return HomeModel{
		BasePageModel: NewBasePageModel("WhatsApp Bot Notify", "home"),
		Description:   "Aplikasi notifikasi WhatsApp",
		Version:       version,
	}
}
