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
	"github.com/gwenziro/bot-notify/internal/server"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
	"github.com/gwenziro/bot-notify/internal/web"
)

func main() {
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

	// Inisialisasi WhatsApp client
	whatsClient, err := client.NewClient(cfg)
	if err != nil {
		utils.Fatal("Gagal inisialisasi WhatsApp client", utils.Fields{"error": err.Error()})
	}
	defer whatsClient.Close()

	// Konfigurasi QR code listener
	whatsClient.SessionManager.SetClient(whatsClient)
	whatsClient.SessionManager.SetupQRCodeListener()

	// Setup handlers
	webHandler := web.NewWebHandler(cfg, whatsClient, nil)
	apiHandler := api.NewAPIHandler(cfg, whatsClient)

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
		utils.Info("Server berjalan", utils.Fields{
			"address": listenAddr,
			"port":    cfg.Server.Port,
			"pid":     os.Getpid(),
		})
		if err := srv.App.Listen(listenAddr); err != nil {
			utils.Error("Error saat menjalankan server", utils.Fields{"error": err.Error()})
		}
	}()

	// Connect WhatsApp dengan sedikit delay untuk memastikan server siap
	go func() {
		time.Sleep(3 * time.Second)
		if err := whatsClient.Connect(); err != nil {
			utils.Error("Gagal terhubung ke WhatsApp", utils.Fields{"error": err.Error()})
		}
	}()

	// Tunggu sinyal shutdown dan tangani graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // Tunggu sinyal shutdown

	utils.Info("Memulai graceful shutdown...")

	// Tutup koneksi WhatsApp dengan bersih
	whatsClient.Disconnect()

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
