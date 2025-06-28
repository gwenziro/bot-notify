package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gwenziro/bot-notify/internal/api"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/manager"
	"github.com/gwenziro/bot-notify/internal/server"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web"
)

func main() {
	// Inisialisasi ProjectRoot sebelum apapun
	utils.InitProjectRoot()

	// Setup dasar logger
	if err := utils.Setup(&utils.LogConfig{Level: "info"}); err != nil {
		fmt.Printf("Error saat inisialisasi logger: %v\n", err)
		os.Exit(1)
	}

	// Log root project dan pastikan struktur direktori
	sysLogger := utils.ForModule("system")
	sysLogger.Info("Detected project root", utils.Fields{"path": utils.ProjectRoot})
	if err := utils.EnsureProjectStructure(); err != nil {
		sysLogger.Error("Gagal membuat struktur direktori", utils.Fields{"error": err.Error()})
		os.Exit(1)
	}

	// Load konfigurasi
	cfg, _ := config.LoadDefault()

	// Setup logger dengan konfigurasi lengkap
	defer utils.Close()

	// Inisialisasi storage
	store, err := storage.Initialize(cfg)
	if err != nil {
		utils.Fatal("Gagal inisialisasi storage", utils.Fields{"error": err.Error()})
	}
	defer store.Close()

	// Inisialisasi UserManager untuk mengelola instance bot multi-user
	userManager := manager.NewUserManager(cfg)
	defer userManager.ShutdownAll()

	// Setup handlers
	webHandler := web.NewWebHandler(cfg, userManager, nil)
	apiHandler := api.NewAPIHandler(cfg, userManager, nil)

	// Buat server dengan template engine yang diaktifkan
	viewsPath := filepath.Join(utils.ProjectRoot, "internal", "web", "view")

	// Tambahkan nilai timeout yang lebih besar di konfigurasi
	cfg.Server.ReadTimeout = 60 * time.Second
	cfg.Server.WriteTimeout = 60 * time.Second

	// Log tambahan untuk memantau loading template
	utils.Info("Memulai inisialisasi web server dengan template", utils.Fields{
		"views_path": viewsPath,
		"exists":     utils.FileExists(viewsPath),
	})

	serverOpts := server.ServerOptions{
		Config:               cfg,
		EnableTemplateEngine: true,
		ViewsPath:            viewsPath,
		WebHandler:           webHandler,
		APIHandler:           apiHandler,
	}

	srv, err := server.NewServer(serverOpts)
	if err != nil {
		utils.Fatal("Gagal inisialisasi server", utils.Fields{"error": err.Error()})
	}

	// Jalankan server di background
	listenAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		utils.Info("Server berjalan dengan sistem multi-user", utils.Fields{
			"address": listenAddr,
			"port":    cfg.Server.Port,
			"pid":     os.Getpid(),
		})
		if err := srv.App.Listen(listenAddr); err != nil {
			utils.Error("Error saat menjalankan server", utils.Fields{"error": err.Error()})
		}
	}()

	// Tunggu sinyal shutdown dan tangani graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // Tunggu sinyal shutdown

	utils.Info("Memulai graceful shutdown...")

	// Shutdown semua instance bot pengguna
	userManager.ShutdownAll()

	// Shutdown HTTP server
	shutdownTimeout := cfg.Server.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 * time.Second
	}

	// Untuk shutdown handler, ganti fiberApp.App dengan fiberApp:
	if err := srv.App.ShutdownWithTimeout(shutdownTimeout); err != nil {
		utils.Error("Error saat shutdown server", utils.Fields{"error": err.Error()})
	} else {
		utils.Info("Server berhasil dimatikan")
	}
}