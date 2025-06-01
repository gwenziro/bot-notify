package entity

// HomePageData merepresentasikan data yang ditampilkan di halaman beranda
type HomePageData struct {
	Title       string // Judul halaman
	Description string // Deskripsi aplikasi
	Version     string // Versi aplikasi
}

// NewHomePageData membuat instance baru HomePageData dengan nilai default
func NewHomePageData() HomePageData {
	return HomePageData{
		Title:       "WhatsApp Bot Notify",
		Description: "Aplikasi notifikasi WhatsApp",
		Version:     "1.0.0",
	}
}
