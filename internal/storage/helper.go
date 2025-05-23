package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Helper adalah utilitas untuk mempermudah operasi penyimpanan
type Helper struct {
	storage Storage
	prefix  string
}

// NewHelper membuat instance baru Helper
func NewHelper(storage Storage, prefix string) *Helper {
	return &Helper{
		storage: storage,
		prefix:  prefix,
	}
}

// makeKey membuat key lengkap dengan prefix
func (h *Helper) makeKey(key string) string {
	return h.prefix + ":" + key
}

// GetJSON mengambil nilai dari key yang diberikan dan unmarshal ke struct
func (h *Helper) GetJSON(ctx context.Context, key string, v interface{}) error {
	data, err := h.storage.Get(ctx, h.makeKey(key))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// SetJSON menyimpan struct dengan key yang diberikan setelah dimarshal ke JSON
func (h *Helper) SetJSON(ctx context.Context, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("gagal marshal data: %w", err)
	}

	return h.storage.Set(ctx, h.makeKey(key), data)
}

// SetJSONWithTTL menyimpan struct dengan key dan TTL yang diberikan
func (h *Helper) SetJSONWithTTL(ctx context.Context, key string, v interface{}, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("gagal marshal data: %w", err)
	}

	return h.storage.SetWithTTL(ctx, h.makeKey(key), data, ttl)
}

// Delete menghapus nilai dengan key yang diberikan
func (h *Helper) Delete(ctx context.Context, key string) error {
	return h.storage.Delete(ctx, h.makeKey(key))
}

// GetAllWithPrefix mengambil semua key-value dengan sub-prefix yang diberikan
func (h *Helper) GetAllWithPrefix(ctx context.Context, subPrefix string) (map[string][]byte, error) {
	return h.storage.GetWithPrefix(ctx, h.makeKey(subPrefix))
}

// DeleteAllWithPrefix menghapus semua key-value dengan sub-prefix yang diberikan
func (h *Helper) DeleteAllWithPrefix(ctx context.Context, subPrefix string) error {
	return h.storage.DeleteWithPrefix(ctx, h.makeKey(subPrefix))
}
