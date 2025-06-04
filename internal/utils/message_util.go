package utils

import (
	"strings"
)

// CleanAndDedupPersonalNumbers membersihkan dan mendeduplikasi nomor personal
func CleanAndDedupPersonalNumbers(numbers []string, logFunc func(msg string, fields Fields)) ([]string, int) {
	cleanedPersonalNumbers := make([]string, 0)
	seenPersonalNumbers := make(map[string]bool)
	dupPersonalCount := 0

	for _, num := range numbers {
		// Lompati nilai kosong/array kosong
		if num == "[]" || num == "[ ]" || num == "" {
			continue
		}

		// Tangani kasus array dalam string
		if strings.HasPrefix(num, "[") && strings.HasSuffix(num, "]") {
			if logFunc != nil {
				logFunc("Memproses array dalam string", Fields{"array": num})
			}

			// Ekstrak nilai-nilai dalam array
			innerArray := ExtractArrayValues(num)
			// Proses hasil ekstraksi dengan deduplikasi
			for _, innerNum := range innerArray {
				if innerNum == "" || !IsValidPhoneNumber(innerNum) {
					continue
				}

				// Normalisasi untuk deduplikasi - menggunakan fungsi dari jid.go
				normalizedNum := FormatPhoneNumber(innerNum)

				if !seenPersonalNumbers[normalizedNum] {
					seenPersonalNumbers[normalizedNum] = true
					cleanedPersonalNumbers = append(cleanedPersonalNumbers, innerNum)
				} else {
					dupPersonalCount++
					if logFunc != nil {
						logFunc("Melewati nomor duplikat", Fields{"number": innerNum})
					}
				}
			}
			continue
		}

		// Nomor normal
		num = strings.TrimSpace(num)
		if num == "" || !IsValidPhoneNumber(num) {
			continue
		}

		// Deduplikasi nomor normal - menggunakan fungsi dari jid.go
		normalizedNum := FormatPhoneNumber(num)
		if !seenPersonalNumbers[normalizedNum] {
			seenPersonalNumbers[normalizedNum] = true
			cleanedPersonalNumbers = append(cleanedPersonalNumbers, num)
		} else {
			dupPersonalCount++
			if logFunc != nil {
				logFunc("Melewati nomor duplikat", Fields{"number": num})
			}
		}
	}

	return cleanedPersonalNumbers, dupPersonalCount
}

// CleanAndDedupGroupIDs membersihkan dan mendeduplikasi ID grup
func CleanAndDedupGroupIDs(ids []string, logFunc func(msg string, fields Fields)) ([]string, int) {
	cleanedGroupIDs := make([]string, 0)
	seenGroupIDs := make(map[string]bool)
	dupGroupCount := 0

	for _, id := range ids {
		// Lompati nilai kosong/array kosong
		if id == "[]" || id == "[ ]" || id == "" {
			continue
		}

		// Tangani kasus array dalam string
		if strings.HasPrefix(id, "[") && strings.HasSuffix(id, "]") {
			if logFunc != nil {
				logFunc("Memproses array dalam string", Fields{"array": id})
			}

			// Ekstrak nilai-nilai dalam array
			innerArray := ExtractArrayValues(id)
			// Proses hasil ekstraksi dengan deduplikasi
			for _, innerID := range innerArray {
				if innerID == "" || !IsValidGroupID(innerID) {
					continue
				}

				// Normalisasi ID grup untuk deduplikasi
				normalizedID := FormatGroupID(innerID)

				if !seenGroupIDs[normalizedID] {
					seenGroupIDs[normalizedID] = true
					cleanedGroupIDs = append(cleanedGroupIDs, innerID)
				} else {
					dupGroupCount++
					if logFunc != nil {
						logFunc("Melewati ID grup duplikat", Fields{"group_id": innerID})
					}
				}
			}
			continue
		}

		// Group ID normal
		id = strings.TrimSpace(id)
		if id == "" || !IsValidGroupID(id) {
			continue
		}

		// Deduplikasi group ID
		normalizedID := FormatGroupID(id)
		if !seenGroupIDs[normalizedID] {
			seenGroupIDs[normalizedID] = true
			cleanedGroupIDs = append(cleanedGroupIDs, id)
		} else {
			dupGroupCount++
			if logFunc != nil {
				logFunc("Melewati ID grup duplikat", Fields{"group_id": id})
			}
		}
	}

	return cleanedGroupIDs, dupGroupCount
}

// ExtractArrayValues mengekstrak nilai-nilai dari string array JSON
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
