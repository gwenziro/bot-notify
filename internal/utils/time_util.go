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

// FormatTimeShort memformat waktu menjadi string pendek
// Format: "26 Mei 2025, 11:35"
func FormatTimeShort(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("02 Jan 2006, 15:04")
}

// FormatTimeIndonesia memformat waktu ke format yang umum di Indonesia
// dengan pemisahan tanggal dan waktu
func FormatTimeIndonesia(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}

	// Format: "27 Mei 2025<br>08:03:10"
	bulanIndonesia := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}

	month := bulanIndonesia[t.Month()-1]
	return fmt.Sprintf("%d %s %d, %02d:%02d:%02d",
		t.Day(), month, t.Year(), t.Hour(), t.Minute(), t.Second())
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

// FormatUptime menghasilkan string uptime yang mudah dibaca
func FormatUptime(duration time.Duration) string {
	// Validasi durasi - batasi untuk mencegah nilai yang tidak masuk akal
	if duration < 0 || duration > 365*24*time.Hour {
		return "Waktu tidak valid"
	}

	seconds := int(duration.Seconds()) % 60
	minutes := int(duration.Minutes()) % 60
	hours := int(duration.Hours()) % 24
	days := int(duration.Hours() / 24)

	// Format yang lebih mudah dibaca
	if days > 0 {
		return fmt.Sprintf("%d hari %d jam %d menit", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d jam %d menit %d detik", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d menit %d detik", minutes, seconds)
	}

	return fmt.Sprintf("%d detik", seconds)
}
