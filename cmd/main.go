package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gwenziro/bot-notify/internal/api"
	"github.com/gwenziro/bot-notify/internal/config"
	"github.com/gwenziro/bot-notify/internal/server"
	"github.com/gwenziro/bot-notify/internal/service/whatsapp/client"
	"github.com/gwenziro/bot-notify/internal/storage"
	"github.com/gwenziro/bot-notify/internal/utils"
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

	// Inisialisasi server Fiber
	fiberApp, err := server.NewFiberApp(cfg)
	if err != nil {
		utils.Fatal("Gagal inisialisasi server", utils.Fields{"error": err.Error()})
	}

	// Setup dan register API handler
	apiHandler := api.NewHandler(cfg, whatsClient)
	apiHandler.RegisterRoutes(fiberApp.App)

	// Jalankan server di background
	listenAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		utils.Info("Server berjalan", utils.Fields{
			"address": listenAddr,
			"port":    cfg.Server.Port,
			"pid":     os.Getpid(),
		})
		if err := fiberApp.App.Listen(listenAddr); err != nil {
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

	if err := fiberApp.App.ShutdownWithTimeout(shutdownTimeout); err != nil {
		utils.Error("Error saat shutdown server", utils.Fields{"error": err.Error()})
	} else {
		utils.Info("Server berhasil dimatikan")
	}
}
