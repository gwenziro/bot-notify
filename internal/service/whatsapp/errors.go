package whatsapp

import "errors"

// Error constants untuk WhatsApp service
var (
	// Connection errors
	ErrWhatsAppNotConnected = errors.New("klien WhatsApp belum terhubung")
	ErrConnectionTimeout    = errors.New("timeout saat menghubungkan ke WhatsApp")
	ErrContextCanceled      = errors.New("konteks dibatalkan saat menunggu koneksi")

	// Data retrieval errors
	ErrGroupNotFound       = errors.New("grup tidak ditemukan")
	ErrContactNotFound     = errors.New("kontak tidak ditemukan")
	ErrSelfIDNotAvailable  = errors.New("ID akun tidak tersedia")
	ErrProfileNotAvailable = errors.New("profil tidak tersedia")
)
