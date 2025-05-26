package utils

import (
	"fmt"
	"time"
)

// FormatTime memformat time.Time menjadi string yang mudah dibaca untuk pengguna Indonesia
// Format: "Senin, 26 Mei 2025 11:35:22"
func FormatTime(t *time.Time) string {
	if t == nil {
		t = &time.Time{}
		*t = time.Now()
	}

	// Tentukan nama hari dalam bahasa Indonesia
	hari := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	namaHari := hari[t.Weekday()]

	// Tentukan nama bulan dalam bahasa Indonesia
	bulan := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	namaBulan := bulan[t.Month()-1]

	// Format: "Senin, 26 Mei 2025 11:35:22"
	return fmt.Sprintf("%s, %02d %s %d %02d:%02d:%02d",
		namaHari, t.Day(), namaBulan, t.Year(),
		t.Hour(), t.Minute(), t.Second())
}

// FormatTimeShort memformat time.Time menjadi string singkat
// Format: "26 Mei 2025, 11:35"
func FormatTimeShort(t *time.Time) string {
	if t == nil {
		t = &time.Time{}
		*t = time.Now()
	}

	// Tentukan nama bulan dalam bahasa Indonesia
	bulan := []string{
		"Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
		"Jul", "Agt", "Sep", "Okt", "Nov", "Des",
	}
	namaBulan := bulan[t.Month()-1]

	// Format: "26 Mei 2025, 11:35"
	return fmt.Sprintf("%02d %s %d, %02d:%02d",
		t.Day(), namaBulan, t.Year(), t.Hour(), t.Minute())
}

// FormatTimeRelative memformat time.Time menjadi waktu relatif
// Misal: "baru saja", "5 menit yang lalu", "1 jam yang lalu", dst
func FormatTimeRelative(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		return "di masa depan" // Jika waktu di masa depan
	}

	// Baru saja: kurang dari 1 menit
	if diff < time.Minute {
		return "baru saja"
	}

	// Dalam menit: 1-59 menit
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d menit yang lalu", minutes)
	}

	// Dalam jam: 1-23 jam
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d jam yang lalu", hours)
	}

	// Dalam hari: 1-6 hari
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d hari yang lalu", days)
	}

	// Dalam minggu: 1-4 minggu
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		return fmt.Sprintf("%d minggu yang lalu", weeks)
	}

	// Dalam bulan: 1-11 bulan
	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		return fmt.Sprintf("%d bulan yang lalu", months)
	}

	// Dalam tahun: >= 1 tahun
	years := int(diff.Hours() / 24 / 365)
	return fmt.Sprintf("%d tahun yang lalu", years)
}
