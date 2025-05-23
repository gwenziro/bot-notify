package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log adalah instance logger global
var Log *logrus.Logger

// LogrusEntry alias untuk entry logrus
type LogrusEntry = *logrus.Entry

// Fields alias untuk fields logrus
type Fields = logrus.Fields

// LogConfig untuk konfigurasi logger
type LogConfig struct {
	Level      string // level log: debug, info, warn, error, fatal, panic
	File       string // path file log (kosong = hanya console)
	MaxSize    int    // ukuran maksimum file log dalam MB
	MaxAge     int    // usia maksimum file log dalam hari
	MaxBackups int    // jumlah backup file log maksimum
	Compress   bool   // kompres file log lama
}

// CleanFormatter adalah formatter kustom yang menghasilkan output rapi
type CleanFormatter struct {
	// TimestampFormat untuk format waktu (default: 15:04:05)
	TimestampFormat string
	// ShowCaller menentukan apakah menampilkan informasi caller
	ShowCaller bool
}

// Format mengimplementasikan logrus.Formatter
func (f *CleanFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 1. Tentukan warna sesuai level dan terapkan ke seluruh baris
	var levelColor, levelName string
	switch entry.Level {
	case logrus.TraceLevel:
		levelColor = "\x1b[37m" // White
		levelName = "TRACE"
	case logrus.DebugLevel:
		levelColor = "\x1b[36m" // Cyan
		levelName = "DEBUG"
	case logrus.InfoLevel:
		levelColor = "\x1b[32m" // Green
		levelName = "INFO "
	case logrus.WarnLevel:
		levelColor = "\x1b[33m" // Yellow
		levelName = "WARN "
	case logrus.ErrorLevel:
		levelColor = "\x1b[31m" // Red
		levelName = "ERROR"
	case logrus.FatalLevel:
		levelColor = "\x1b[35m" // Magenta
		levelName = "FATAL"
	case logrus.PanicLevel:
		levelColor = "\x1b[35m" // Magenta
		levelName = "PANIC"
	}

	reset := "\x1b[0m"

	// 2. Format timestamp
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = "15:04:05"
	}
	timestamp := entry.Time.Format(timestampFormat)

	// 3. Format module lebih menonjol dan hapus bagian caller file:line
	module := "system"
	if mod, ok := entry.Data["module"]; ok {
		module = fmt.Sprintf("%s", mod)
		delete(entry.Data, "module")
	}

	// Hilangkan caller information karena kita menggunakan module

	// 4. Mulai dengan warna level
	fmt.Fprint(b, levelColor)

	// 5. Format pesan utama dengan warna sesuai level dan module sebagai identifikasi utama
	fmt.Fprintf(b, "%s %s [%s] » %s", timestamp, levelName, module, entry.Message)

	// 6. Tambahkan field lain dengan warna yang sama tetapi dengan highlight berbeda untuk key
	if len(entry.Data) > 0 {
		fieldSeparator := " "
		for k, v := range entry.Data {
			var value string
			switch val := v.(type) {
			case string:
				value = val
			case error:
				value = val.Error()
			default:
				value = fmt.Sprintf("%+v", val)
			}

			// Key dengan warna bold dan value dengan warna normal
			fmt.Fprintf(b, "%s\x1b[1m%s\x1b[0m%s=%s", fieldSeparator, k, levelColor, value)
			fieldSeparator = " "
		}
	}

	// 7. Akhiri dengan reset warna dan newline
	fmt.Fprint(b, reset)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

// Setup menginisialisasi logger
func Setup(cfg *LogConfig) error {
	// Buat logger baru jika belum ada
	if Log == nil {
		Log = logrus.New()
	}

	// Setup formatter yang simpel dan berwarna
	Log.SetFormatter(&CleanFormatter{
		TimestampFormat: "15:04:05",
		ShowCaller:      false, // Matikan caller info karena kita menggunakan module
	})

	// Set level log
	level := logrus.InfoLevel // Default level
	if cfg != nil && cfg.Level != "" {
		if lvl, err := logrus.ParseLevel(cfg.Level); err == nil {
			level = lvl
		}
	}
	Log.SetLevel(level)

	// Non-aktifkan caller reporting karena kita menggunakan module
	Log.SetReportCaller(false)

	// Konfigurasi output ke file jika diperlukan
	if cfg != nil && cfg.File != "" {
		// Pastikan direktori log ada
		logDir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// Setup file rotasi log
		fileLogger := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		// Output ke file dan konsol
		Log.SetOutput(io.MultiWriter(os.Stdout, fileLogger))
	} else {
		// Output hanya ke konsol
		Log.SetOutput(os.Stdout)
	}

	Info("Logger berhasil diinisialisasi", Fields{"level": level.String()})
	return nil
}

// ForModule mengembalikan logger dengan nama modul yang konsisten
func ForModule(module string) LogrusEntry {
	if Log == nil {
		// Inisialisasi logger minimal jika belum ada
		Setup(nil)
	}

	// Standardisasi nama modul untuk konsistensi
	module = strings.ToLower(module)

	return Log.WithField("module", module)
}

// Debug logs pesan pada level Debug
func Debug(msg string, fields ...Fields) {
	if Log == nil {
		Setup(nil)
	}

	if len(fields) > 0 {
		Log.WithFields(fields[0]).Debug(msg)
	} else {
		Log.Debug(msg)
	}
}

// Info logs pesan pada level Info
func Info(msg string, fields ...Fields) {
	if Log == nil {
		Setup(nil)
	}

	if len(fields) > 0 {
		Log.WithFields(fields[0]).Info(msg)
	} else {
		Log.Info(msg)
	}
}

// Warn logs pesan pada level Warn
func Warn(msg string, fields ...Fields) {
	if Log == nil {
		Setup(nil)
	}

	if len(fields) > 0 {
		Log.WithFields(fields[0]).Warn(msg)
	} else {
		Log.Warn(msg)
	}
}

// Error logs pesan pada level Error
func Error(msg string, fields ...Fields) {
	if Log == nil {
		Setup(nil)
	}

	if len(fields) > 0 {
		Log.WithFields(fields[0]).Error(msg)
	} else {
		Log.Error(msg)
	}
}

// Fatal logs pesan pada level Fatal kemudian exit dengan status 1
func Fatal(msg string, fields ...Fields) {
	if Log == nil {
		Setup(nil)
	}

	if len(fields) > 0 {
		Log.WithFields(fields[0]).Fatal(msg)
	} else {
		Log.Fatal(msg)
	}
}

// Close melakukan clean-up jika diperlukan
func Close() {
	// Tidak diperlukan implementasi khusus karena logrus
	// akan otomatis flush ke output
}
