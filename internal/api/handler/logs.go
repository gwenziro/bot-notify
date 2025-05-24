package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/service/log"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// LogsHandler menangani API request untuk logs
type LogsHandler struct {
	logService *log.LogService
	logger     utils.LogrusEntry
}

// NewLogsHandler membuat instance baru logs handler
func NewLogsHandler(logService *log.LogService, logger utils.LogrusEntry) *LogsHandler {
	return &LogsHandler{
		logService: logService,
		logger:     logger.WithField("component", "logs-handler"),
	}
}

// GetLogs mengambil data logs dengan paginasi dan filter
func (h *LogsHandler) GetLogs(c *fiber.Ctx) error {
	// Parameter paginasi dan filter
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 25)
	level := c.Query("level", "")
	source := c.Query("source", "")
	search := c.Query("search", "")
	from := c.Query("from", "")
	to := c.Query("to", "")

	// Log request dengan fields yang dipakai
	h.logger.WithFields(utils.Fields{
		"page":   page,
		"limit":  limit,
		"level":  level,
		"source": source,
		"search": search,
		"from":   from,
		"to":     to,
	}).Debug("API request: GetLogs")

	// Ambil logs dari service
	logs, totalLogs, totalPages, err := h.logService.GetPaginatedLogs(page, limit, level, source, search, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Gagal mendapatkan logs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Gagal mengambil logs",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"logs":       logs,
		"totalLogs":  totalLogs,
		"totalPages": totalPages,
		"page":       page,
		"limit":      limit,
		"success":    true,
	})
}

// ClearLogs menghapus semua logs
func (h *LogsHandler) ClearLogs(c *fiber.Ctx) error {
	h.logger.Info("API request: ClearLogs")

	err := h.logService.ClearAllLogs()
	if err != nil {
		h.logger.WithError(err).Error("Gagal menghapus logs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Gagal menghapus logs",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Semua logs berhasil dihapus",
	})
}

// ExportLogs mengekspor logs dalam format tertentu (CSV atau JSON)
func (h *LogsHandler) ExportLogs(c *fiber.Ctx) error {
	format := c.Query("format", "csv")
	level := c.Query("level", "")
	source := c.Query("source", "")
	search := c.Query("search", "")
	from := c.Query("from", "")
	to := c.Query("to", "")

	h.logger.WithFields(utils.Fields{
		"format": format,
		"level":  level,
		"source": source,
	}).Info("API request: ExportLogs")

	// Panggil service untuk ekspor
	fileBytes, fileName, contentType, err := h.logService.ExportLogs(format, level, source, search, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Gagal mengekspor logs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Gagal mengekspor logs",
		})
	}

	// Set header untuk download
	c.Set("Content-Disposition", "attachment; filename="+fileName)
	c.Set("Content-Type", contentType)

	return c.Send(fileBytes)
}

// GetLogSummary mendapatkan ringkasan log untuk dashboard
func (h *LogsHandler) GetLogSummary(c *fiber.Ctx) error {
	h.logger.Debug("API request: GetLogSummary")

	// Implementasi sederhana untuk demonstrasi
	// Di produksi, Anda harus mengambil data aktual
	summary := fiber.Map{
		"total_logs":    0,
		"error_count":   0,
		"warning_count": 0,
		"today_logs":    0,
		"latest": []fiber.Map{
			{
				"timestamp": "2023-09-10T08:45:23Z",
				"level":     "INFO",
				"message":   "Sistem dimulai",
			},
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"summary": summary,
	})
}
