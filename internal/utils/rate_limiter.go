package utils

import (
	"sync"
	"time"
)

// RateLimiter adalah helper untuk membatasi jumlah request berdasarkan IP
type RateLimiter struct {
	rateMap     map[string][]time.Time
	maxRequests int
	interval    time.Duration
	mutex       sync.RWMutex
}

// NewRateLimiter membuat RateLimiter baru
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		rateMap:     make(map[string][]time.Time),
		maxRequests: maxRequests,
		interval:    interval,
		mutex:       sync.RWMutex{},
	}
}

// AddRequest menambahkan request untuk IP tertentu
func (r *RateLimiter) AddRequest(ip string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Hapus timestamp yang sudah berlalu
	now := time.Now()
	cutoff := now.Add(-r.interval)

	if timestamps, exists := r.rateMap[ip]; exists {
		// Filter timestamp yang masih dalam window
		newTimestamps := []time.Time{}
		for _, t := range timestamps {
			if t.After(cutoff) {
				newTimestamps = append(newTimestamps, t)
			}
		}

		// Tambahkan timestamp baru
		r.rateMap[ip] = append(newTimestamps, now)
	} else {
		// Jika IP baru, init dengan timestamp saat ini
		r.rateMap[ip] = []time.Time{now}
	}

	// Cleanup logic
	if len(r.rateMap) > 1000 {
		r.cleanup(cutoff)
	}
}

// ExceedsLimit memeriksa apakah IP melebihi batas
func (r *RateLimiter) ExceedsLimit(ip string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Hapus timestamp yang sudah berlalu
	now := time.Now()
	cutoff := now.Add(-r.interval)

	// Hitung request yang valid dalam interval
	var validRequests int

	if timestamps, exists := r.rateMap[ip]; exists {
		for _, t := range timestamps {
			if t.After(cutoff) {
				validRequests++
			}
		}
	}

	// Bandingkan dengan max
	return validRequests >= r.maxRequests
}

// GetRemainingRequests mendapatkan sisa request yang diizinkan
func (r *RateLimiter) GetRemainingRequests(ip string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Hapus timestamp yang sudah berlalu
	now := time.Now()
	cutoff := now.Add(-r.interval)

	// Hitung request yang valid dalam interval
	var validRequests int

	if timestamps, exists := r.rateMap[ip]; exists {
		for _, t := range timestamps {
			if t.After(cutoff) {
				validRequests++
			}
		}
	}

	// Hitung sisa
	remaining := r.maxRequests - validRequests
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// Reset menghapus semua data rate limit
func (r *RateLimiter) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.rateMap = make(map[string][]time.Time)
}

// ResetIP menghapus data rate limit untuk IP tertentu
func (r *RateLimiter) ResetIP(ip string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.rateMap, ip)
}

// cleanup menghapus IP yang sudah tidak aktif
func (r *RateLimiter) cleanup(cutoff time.Time) {
	for ip, timestamps := range r.rateMap {
		if len(timestamps) == 0 {
			delete(r.rateMap, ip)
			continue
		}

		// Periksa timestamp terakhir
		lastTimestamp := timestamps[len(timestamps)-1]
		if lastTimestamp.Before(cutoff) {
			delete(r.rateMap, ip)
		}
	}
}
