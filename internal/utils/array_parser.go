package utils

import (
	"strings"
)

// ExtractArrayValues mengekstrak nilai-nilai dari string array JSON
// Contoh: "[\"value1\", \"value2\"]" -> []string{"value1", "value2"}
func ExtractArrayValues(arrayStr string) []string {
	// Hapus tanda bracket di awal dan akhir
	content := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(arrayStr, "]"), "["))
	if content == "" {
		return []string{}
	}

	// Split berdasarkan koma tetapi hitung tanda kutip
	var result []string
	var currentValue strings.Builder
	inQuotes := false

	for _, char := range content {
		switch char {
		case '"', '\'':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				// Koma di luar tanda kutip = pemisah
				val := strings.TrimSpace(currentValue.String())
				// Bersihkan tanda kutip
				val = strings.Trim(val, `"'`)
				result = append(result, val)
				currentValue.Reset()
				continue
			}
		}
		currentValue.WriteRune(char)
	}

	// Tambahkan nilai terakhir jika ada
	if currentValue.Len() > 0 {
		val := strings.TrimSpace(currentValue.String())
		val = strings.Trim(val, `"'`)
		result = append(result, val)
	}

	return result
}

// IsArrayString memeriksa apakah string merupakan representasi array JSON
func IsArrayString(str string) bool {
	trimmed := strings.TrimSpace(str)
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

// IsEmptyArrayString memeriksa apakah string merupakan representasi array kosong
func IsEmptyArrayString(str string) bool {
	trimmed := strings.TrimSpace(str)
	return trimmed == "[]" || trimmed == "[ ]"
}
