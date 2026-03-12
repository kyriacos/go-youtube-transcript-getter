package transcript

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var reSanitize = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizeKey(key string) string {
	return reSanitize.ReplaceAllString(key, "_")
}

type fsCacheItem struct {
	Value      string `json:"value"`
	Expiration int64  `json:"expiration"`
}

type FsCache struct {
	dir string
}

func NewFsCache(dir string) (*FsCache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FsCache{dir: dir}, nil
}

func (c *FsCache) Get(ctx context.Context, key string) (string, error) {
	path := filepath.Join(c.dir, sanitizeKey(key))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var item fsCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return "", err
	}

	if item.Expiration != 0 && time.Now().UnixMilli() > item.Expiration {
		_ = os.Remove(path)
		return "", nil
	}

	return item.Value, nil
}

func (c *FsCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	var expiration int64
	if ttl != 0 {
		expiration = time.Now().Add(ttl).UnixMilli()
	}

	item := fsCacheItem{
		Value:      value,
		Expiration: expiration,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	path := filepath.Join(c.dir, sanitizeKey(key))
	return os.WriteFile(path, data, 0644)
}
