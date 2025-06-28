# Bot Notify - Multi-User WhatsApp Bot System

Bot Notify adalah sistem bot WhatsApp yang mendukung multi-user dengan instance bot yang terpisah untuk setiap pengguna. Setiap pengguna memiliki bot WhatsApp yang berjalan secara independen dengan data dan state yang terisolasi.

## Fitur Multi-User

### 🔐 Isolasi Pengguna
- **Instance Bot Terpisah**: Setiap pengguna memiliki instance `whatsmeow.Client` yang independen
- **Data Terisolasi**: Session, QR code, dan storage disimpan dalam direktori terpisah per pengguna
- **Concurrent Safe**: Sistem dapat menangani permintaan dari multiple pengguna secara bersamaan

### 🏗️ Arsitektur Sistem

#### UserManager
Komponen utama yang mengelola semua instance bot pengguna:

```go
type UserManager struct {
    clients      map[string]*client.Client  // userID -> client instance
    mutex        sync.RWMutex               // Thread-safe access
    globalConfig *config.Config             // Template konfigurasi
    logger       utils.LogrusEntry          // Logging
}
```

#### Struktur Direktori Per Pengguna
```
data/
└── users/
    ├── user123/
    │   ├── whatsapp/     # Session WhatsApp
    │   ├── qrcodes/      # QR codes
    │   ├── sessions/     # Web sessions
    │   └── storage/      # Database storage
    └── admin/
        ├── whatsapp/
        ├── qrcodes/
        ├── sessions/
        └── storage/
```

### 🔑 Sistem Autentikasi

#### Token Admin
- Token global untuk akses admin: `X-Access-Token: your-admin-token`
- Memberikan akses ke semua fitur termasuk endpoint admin

#### Token Pengguna
- UserID sebagai token: `X-Access-Token: user123`
- Setiap userID otomatis membuat instance bot baru jika belum ada
- Format userID: 3-50 karakter alphanumeric, underscore, dan dash

### 📡 API Endpoints

#### Endpoint Umum
```bash
# Status koneksi pengguna tertentu
GET /api/status
Headers: X-Access-Token: user123

# Kirim pesan personal
POST /api/send/personal
Headers: X-Access-Token: user123
Body: {"phoneNumber": "628123456789", "message": "Hello"}

# Kirim pesan grup
POST /api/send/group
Headers: X-Access-Token: user123
Body: {"groupID": "120363123456789@g.us", "message": "Hello Group"}
```

#### Endpoint Admin
```bash
# Status semua pengguna (hanya admin)
GET /api/admin/users/status
Headers: X-Access-Token: your-admin-token
```

### 🔄 Lifecycle Management

#### Pembuatan Instance
```go
// Otomatis saat request pertama dengan userID baru
userClient, err := userManager.NewUserClient("user123")
```

#### Cleanup Otomatis
- **Periodic Cleanup**: Setiap 30 menit, sistem membersihkan client yang tidak aktif > 2 jam
- **Graceful Shutdown**: Semua instance ditutup dengan bersih saat aplikasi dimatikan
- **Resource Management**: Memory dan file handles dibersihkan secara otomatis

### 🛡️ Keamanan & Isolasi

#### Data Isolation
- Setiap pengguna memiliki database SQLite terpisah
- Session WhatsApp disimpan dalam direktori terpisah
- QR codes tidak dapat diakses antar pengguna

#### Concurrent Safety
- Semua operasi pada `UserManager` menggunakan `sync.RWMutex`
- Thread-safe access ke client instances
- Tidak ada shared state antar pengguna

#### Error Handling
- Client yang error tidak mempengaruhi client lain
- Automatic retry dan reconnection per instance
- Comprehensive logging per pengguna

### 📊 Monitoring & Logging

#### Per-User Logging
```go
logger := utils.ForModule(fmt.Sprintf("client-%s", userID))
```

#### Metrics
- Jumlah pengguna aktif: `userManager.GetUserCount()`
- Status semua pengguna: `userManager.GetAllUsers()`
- Health check per instance

### 🚀 Deployment

#### Environment Variables
```bash
# Konfigurasi tetap sama, path akan diisolasi otomatis
WHATSAPP_STORE_DIR=./data/whatsapp
WHATSAPP_QR_CODE_DIR=./data/qrcodes
AUTH_SESSION_DIR=./data/sessions
```

#### Scaling Considerations
- **Memory Usage**: ~10-50MB per active user instance
- **File Descriptors**: ~5-10 per user instance
- **Database Connections**: 1 SQLite connection per user
- **Recommended Limit**: 100-500 concurrent users per server

### 🔧 Configuration

#### User-Specific Paths
Sistem otomatis membuat path yang terisolasi:
```go
userConfig.WhatsApp.StoreDir = "data/users/{userID}/whatsapp"
userConfig.WhatsApp.QrCodeDir = "data/users/{userID}/qrcodes"
userConfig.Auth.SessionDir = "data/users/{userID}/sessions"
userConfig.Storage.Path = "data/users/{userID}/storage"
```

#### Health Checks
- Connection health check per user instance
- Automatic reconnection pada connection loss
- Periodic cleanup inactive instances

### 📝 Usage Examples

#### Mengirim Pesan sebagai User
```bash
curl -X POST "http://localhost:8080/api/send/personal" \
  -H "X-Access-Token: user123" \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "628123456789",
    "message": "Hello from user123 bot!"
  }'
```

#### Monitoring sebagai Admin
```bash
curl -X GET "http://localhost:8080/api/admin/users/status" \
  -H "X-Access-Token: your-admin-token"
```

### 🔄 Migration dari Single-User

Sistem ini backward compatible dengan konfigurasi single-user. Admin client menggunakan userID "admin" dan dapat diakses melalui dashboard web seperti sebelumnya.

### 🎯 Best Practices

1. **UserID Naming**: Gunakan identifier yang konsisten (email, username, UUID)
2. **Resource Limits**: Monitor memory usage dan set limits sesuai server capacity
3. **Cleanup Strategy**: Sesuaikan interval cleanup berdasarkan usage pattern
4. **Monitoring**: Implement proper monitoring untuk track active users dan resource usage
5. **Backup**: Backup direktori `data/users/` untuk preserve user sessions

Sistem multi-user ini memungkinkan skalabilitas horizontal yang baik sambil mempertahankan isolasi dan keamanan data antar pengguna.