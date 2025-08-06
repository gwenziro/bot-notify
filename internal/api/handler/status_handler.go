package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/api/middleware"
	"github.com/gwenziro/bot-notify/internal/api/model"
	"github.com/gwenziro/bot-notify/internal/constants"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// StatusHandler menangani endpoint status koneksi API dengan dukungan multi-user
type StatusHandler struct {
	BaseHandler
	version string // Versi aplikasi untuk ping response
}

// NewStatusHandler membuat instance baru StatusHandler
func NewStatusHandler(userManager *manager.UserManager) *StatusHandler {
	return &StatusHandler{
		BaseHandler: NewBaseHandler(userManager, "handler-status"),
		version:     constants.DefaultAppVersion,
	}
}

// GetStatus mengembalikan status koneksi WhatsApp untuk pengguna tertentu
// Endpoint: GET /api/status
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetStatus")

	// 2. Dapatkan client untuk user ini
	whatsClient, userID, err := h.GetUserClient(c)
	if err != nil {
		return h.SendError(c, constants.MsgClientNotAvailable, err, fiber.StatusInternalServerError)
	}

	// 3. Periksa ketersediaan WhatsApp client
	if whatsClient == nil {
		return h.SendError(c, constants.MsgClientNotAvailable, nil, fiber.StatusInternalServerError)
	}

	// 4. Dapatkan status koneksi
	state, err := whatsClient.GetConnectionStateSafe()
	if err != nil {
		return h.SendError(c, constants.MsgStatusFailed, err, fiber.StatusInternalServerError)
	}

	// 5. Tentukan pesan yang sesuai
	var message string
	if !state.IsConnected {
		message = constants.MsgNotConnected
	} else {
		message = constants.MsgConnected
	}

	// 6. Buat respons
	response := model.NewStatusResponse(message, string(state.Status), state.IsConnected)

	// 7. Sertakan informasi tambahan jika terhubung
	if state.IsConnected {
		response.LastActivity = utils.FormatTimeIndonesia(&state.LastActivity)
	}

	// 8. Selalu kembalikan StatusOK (200) agar dashboard tetap bisa menampilkan status
	statusCode := fiber.StatusOK

	// 9. Log hasil
	h.LogSuccessResponse("Status koneksi berhasil diambil", utils.Fields{
		"userID":      userID,
		"status":      state.Status,
		"connected":   state.IsConnected,
		"http_status": statusCode,
	})

	return c.Status(statusCode).JSON(response)
}

// TestConnection menguji koneksi API tanpa autentikasi
// Endpoint: GET /ping
func (h *StatusHandler) TestConnection(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "TestConnection")

	// 2. Buat respons ping
	pingResponse := model.NewPingResponse("API berfungsi dengan baik", h.version)

	// 3. Log hasil
	h.LogSuccessResponse("Test koneksi berhasil", utils.Fields{
		"version": h.version,
	})

	// 4. Kirim respons
	return h.SendSuccess(c, pingResponse)
}

// GetAllUsersStatus mengembalikan status semua pengguna (hanya untuk admin)
// Endpoint: GET /api/admin/users/status
func (h *StatusHandler) GetAllUsersStatus(c *fiber.Ctx) error {
	// 1. Log informasi debug request
	h.LogDebugRequest(c, "GetAllUsersStatus")

	// 2. Cek apakah user adalah admin
	if !middleware.IsAdminFromContext(c) {
		return h.SendError(c, "Akses ditolak: hanya admin yang dapat melihat status semua pengguna", nil, fiber.StatusForbidden)
	}

	// 3. Dapatkan informasi semua pengguna
	allUsers := h.UserManager.GetAllUsers()

	// 4. Buat respons
	response := map[string]interface{}{
		"success":    true,
		"message":    "Status semua pengguna berhasil diambil",
		"time":       utils.FormatTimeIndonesia(&[]time.Time{time.Now()}[0]),
		"totalUsers": len(allUsers),
		"users":      allUsers,
	}

	// 5. Log hasil
	h.LogSuccessResponse("Status semua pengguna berhasil diambil", utils.Fields{
		"totalUsers": len(allUsers),
	})

	return h.SendSuccess(c, response)
}