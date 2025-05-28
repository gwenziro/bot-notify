package constants

// Status koneksi
const (
	MsgConnected    = "WhatsApp terhubung dan siap digunakan"
	MsgNotConnected = "WhatsApp sedang tidak terhubung"
	MsgConnecting   = "WhatsApp sedang dalam proses koneksi"
	MsgLoggedOut    = "WhatsApp telah logout, silakan login kembali"
)

// Pesan error umum
const (
	MsgServerError        = "Terjadi kesalahan pada server"
	MsgInvalidRequest     = "Format request tidak valid"
	MsgInvalidParam       = "Parameter tidak valid"
	MsgMissingField       = "Field %s tidak boleh kosong"
	MsgClientNotAvailable = "WhatsApp client tidak tersedia"
	MsgStatusFailed       = "Gagal mendapatkan status koneksi: %v"
)

// Pesan terkait koneksi
const (
	MsgConnectionFailed    = "Gagal menghubungkan WhatsApp"
	MsgReconnectSuccess    = "Permintaan menghubungkan ulang WhatsApp berhasil diproses"
	MsgDisconnectSuccess   = "WhatsApp berhasil diputuskan dan sesi dibersihkan"
	MsgFailedSessionDelete = "Koneksi diputus tetapi gagal menghapus sesi"
)

// Pesan terkait QR code
const (
	MsgQrNotAvailable = "QR code belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu"
	MsgQrExpired      = "QR code sudah kedaluwarsa. Silakan gunakan endpoint /api/reconnect untuk mendapatkan QR code baru"
	MsgQrAvailable    = "QR code tersedia, silakan pindai"
	MsgQrConnected    = "QR code tidak tersedia: WhatsApp sudah terhubung"
	MsgQrNotConnected = "QR code tidak tersedia: WhatsApp sedang tidak terhubung"
)

// Pesan terkait pengiriman pesan
const (
	MsgSendSuccess        = "Notifikasi WhatsApp terkirim!"
	MsgSendGroupSuccess   = "Notifikasi WhatsApp terkirim ke grup!"
	MsgSendFailure        = "Gagal mengirim pesan"
	MsgBroadcastSuccess   = "Broadcast berhasil diproses"
	MsgMinTarget          = "Minimal harus ada 2 target penerima untuk broadcast"
	MsgMaxTarget          = "Maksimal hanya 16 target penerima untuk broadcast"
	MsgEmptyMessage       = "Pesan tidak boleh kosong"
	MsgInvalidPhoneNumber = "Format nomor telepon tidak valid"
	MsgTimeoutError       = "Timeout saat mengirim pesan: operasi terlalu lama"
)

// Pesan terkait grup
const (
	MsgGroupsRetrieved            = "Daftar grup berhasil diambil"
	MsgGroupParticipantsRetrieved = "Daftar anggota grup berhasil diambil"
	MsgGroupNotFound              = "Grup tidak ditemukan"
	MsgInvalidGroupID             = "Format ID grup tidak valid"
	MsgGroupIDRequired            = "ID grup harus disediakan"
	MsgFindIDFailed               = "Gagal mendapatkan ID %s"
	MsgGroupDataRetrievalFailed   = "Gagal mendapatkan data grup: %v"
)

// Pesan terkait profil
const (
	MsgProfileRetrieved = "Profil WhatsApp berhasil diambil"
)
