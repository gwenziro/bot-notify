package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/service/log"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/utils"
)

// LogsController menangani tampilan untuk halaman logs
type LogsController struct {
	config     *config.Config
	whatsapp   *client.Client
	logger     utils.LogrusEntry
	logService *log.LogService
}

// NewLogsController membuat instance baru logs controller
func NewLogsController(cfg *config.Config, whatsapp *client.Client, logService *log.LogService, logger utils.LogrusEntry) *LogsController {
	return &LogsController{
		config:     cfg,
		whatsapp:   whatsapp,
		logger:     logger.WithField("component", "logs-controller"),
		logService: logService,
	}
}

// LogsPage menampilkan halaman log sistem
func (c *LogsController) LogsPage(ctx *fiber.Ctx) error {
	// Log akses ke halaman logs
	c.logger.Info("Mengakses halaman logs")

	// Render dengan layout dashboard
	return ctx.Render("dashboard/logs", fiber.Map{
		"Title":       "Log Sistem",
		"Description": "Monitoring dan ekspor log sistem Bot Notify.",
		"ActivePage":  "logs", // Untuk highlight menu aktif di sidebar
		"User":        ctx.Locals("user"),
	}, "layouts/dashboard")
}

// GetServerSideData digunakan untuk mendapatkan data logs untuk server-side rendering
func (c *LogsController) GetServerSideData(ctx *fiber.Ctx) error {
	// Parameter paginasi dan filter
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 25)
	level := ctx.Query("level", "")
	source := ctx.Query("source", "")
	search := ctx.Query("search", "")
	from := ctx.Query("from", "")
	to := ctx.Query("to", "")

	// Ambil data logs
	logs, totalLogs, totalPages, err := c.logService.GetPaginatedLogs(page, limit, level, source, search, from, to)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mendapatkan data logs")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data logs",
		})
	}

	// Return data dalam format JSON untuk tabel
	return ctx.JSON(fiber.Map{
		"logs":       logs,
		"total":      totalLogs,
		"totalPages": totalPages,
		"page":       page,
		"limit":      limit,
		"success":    true,
	})
}

// ClearLogs menghapus semua log dari sistem
func (c *LogsController) ClearLogs(ctx *fiber.Ctx) error {
	if err := c.logService.ClearAllLogs(); err != nil {
		c.logger.WithError(err).Error("Gagal menghapus semua logs")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus log",
		})
	}

	c.logger.Info("Semua log berhasil dihapus")
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Semua log berhasil dihapus",
	})
}

// ExportLogs mengekspor log untuk diunduh
func (c *LogsController) ExportLogs(ctx *fiber.Ctx) error {
	format := ctx.Query("format", "csv")
	level := ctx.Query("level", "")
	source := ctx.Query("source", "")
	search := ctx.Query("search", "")
	from := ctx.Query("from", "")
	to := ctx.Query("to", "")

	// Generate file export
	fileBytes, fileName, contentType, err := c.logService.ExportLogs(format, level, source, search, from, to)
	if err != nil {
		c.logger.WithError(err).Error("Gagal mengekspor logs")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengekspor logs",
		})
	}

	// Set header untuk download
	ctx.Set("Content-Disposition", "attachment; filename="+fileName)
	ctx.Set("Content-Type", contentType)

	return ctx.Send(fileBytes)
}

// GetLogLevels mengembalikan daftar level log yang tersedia
func (c *LogsController) GetLogLevels(ctx *fiber.Ctx) error {
	levels := []string{"INFO", "DEBUG", "WARNING", "ERROR", "FATAL"}
	return ctx.JSON(fiber.Map{
		"levels": levels,
	})
}

// GetLogSources mengembalikan daftar sumber log yang tersedia
func (c *LogsController) GetLogSources(ctx *fiber.Ctx) error {
	sources := []string{"SYSTEM", "WHATSAPP", "API", "WEB", "DATABASE"}
	return ctx.JSON(fiber.Map{
		"sources": sources,
	})
}
