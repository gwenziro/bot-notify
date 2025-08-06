package entity

// HomePageData merepresentasikan data yang ditampilkan di halaman beranda
type HomePageData struct {
	Title       string // Judul halaman
	Description string // Deskripsi aplikasi
	Version     string // Versi aplikasi
	CurrentYear int    // Tahun saat ini untuk footer
	ApiBaseURL  string // Base URL API untuk contoh kode
}

// NewHomePageData membuat instance baru HomePageData dengan nilai default
func NewHomePageData() HomePageData {
	return HomePageData{
		Title:       "WhatsApp Bot Notify",
		Description: "Bot WhatsApp Kirim Pesan Realtime",
		Version:     "1.0.0",
		CurrentYear: 0,
		ApiBaseURL:  "http://localhost:8080",
	}
}
