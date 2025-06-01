package utils

// GetFirstOrEmpty mengambil elemen pertama array atau string kosong jika array kosong
func GetFirstOrEmpty(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}
