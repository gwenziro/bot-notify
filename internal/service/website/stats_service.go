package website // Ubah dari "webservice" ke "website"

// StatisticsService mengelola statistik aplikasi
type StatisticsService struct {
	messagesSent int
	// metrik lainnya...
}

// NewStatisticsService membuat dan mengembalikan instance baru dari StatisticsService
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

// IncrementMessageCount menambah jumlah pesan terkirim
func (s *StatisticsService) IncrementMessageCount() {
	s.messagesSent++
}

// GetMessageCount mengembalikan jumlah pesan terkirim
func (s *StatisticsService) GetMessageCount() int {
	return s.messagesSent
}
