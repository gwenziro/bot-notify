package constants

// Konstanta untuk status koneksi
const (
	MsgConnected    = "WhatsApp terhubung dan siap digunakan"
	MsgNotConnected = "WhatsApp sedang tidak terhubung"
	MsgConnecting   = "WhatsApp sedang dalam proses koneksi"
	MsgLoggedOut    = "WhatsApp telah logout, silakan login kembali"
)

// Konstanta untuk pesan error umum
const (
	MsgServerError        = "Terjadi kesalahan pada server"
	MsgInvalidRequest     = "Format request tidak valid"
	MsgInvalidParam       = "Parameter tidak valid"
	MsgMissingField       = "Field %s tidak boleh kosong"
	MsgClientNotAvailable = "WhatsApp client tidak tersedia"
	MsgStatusFailed       = "Gagal mendapatkan status koneksi: %v"
	MsgTimeoutError       = "Operasi melebihi batas waktu yang ditentukan"
)

// Konstanta untuk pesan koneksi
const (
	MsgConnectionFailed    = "Gagal menghubungkan WhatsApp"
	MsgReconnectSuccess    = "Permintaan menghubungkan ulang WhatsApp berhasil diproses"
	MsgDisconnectSuccess   = "WhatsApp berhasil diputuskan dan sesi dibersihkan"
	MsgFailedSessionDelete = "Koneksi diputus tetapi gagal menghapus sesi"
)

// Konstanta untuk QR code
const (
	MsgQrNotAvailable = "QR code belum tersedia. Silakan gunakan endpoint /api/reconnect terlebih dahulu"
	MsgQrExpired      = "QR code sudah kedaluwarsa. Silakan gunakan endpoint /api/reconnect untuk mendapatkan QR code baru"
	MsgQrAvailable    = "QR code tersedia, silakan pindai"
	MsgQrConnected    = "QR code tidak tersedia: WhatsApp sudah terhubung"
	MsgQrNotConnected = "QR code tidak tersedia: WhatsApp sedang tidak terhubung"
)

// Konstanta untuk pengiriman pesan
const (
	MsgSendSuccess        = "Notifikasi WhatsApp terkirim!"
	MsgSendGroupSuccess   = "Notifikasi WhatsApp terkirim ke grup!"
	MsgSendFailure        = "Gagal mengirim pesan"
	MsgInvalidPhoneNumber = "Format nomor telepon tidak valid"
	MsgInvalidGroupID     = "Format ID grup tidak valid"
)

// Konstanta untuk operasi broadcast
const (
	MsgBroadcastSuccess         = "Broadcast berhasil diproses"
	MsgMinTarget                = "Minimal harus ada 2 target penerima untuk broadcast"
	MsgMaxTarget                = "Maksimal hanya 16 target penerima untuk broadcast"
	MsgEmptyMessage             = "Pesan tidak boleh kosong"
	MsgBroadcastConnectionError = "Gagal mengirim pesan broadcast: WhatsApp sedang tidak terhubung"
)

// Konstanta untuk grup
const (
	MsgGroupsRetrieved            = "Daftar grup berhasil diambil"
	MsgGroupParticipantsRetrieved = "Daftar anggota grup berhasil diambil"
	MsgGroupNotFound              = "Grup tidak ditemukan"
	MsgGroupIDRequired            = "ID grup harus disediakan"
	MsgFindIDFailed               = "Gagal mendapatkan ID %s"
	MsgGroupDataRetrievalFailed   = "Gagal mendapatkan data grup: %v"
)

// Konstanta untuk profil
const (
	MsgProfileRetrieved = "Profil WhatsApp berhasil diambil"
)
