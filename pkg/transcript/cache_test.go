package transcript

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryCache(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache()

	t.Run("Store and retrieve", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		got, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Fatal(err)
		}
		if got != "value1" {
			t.Errorf("Expected value1, got %s", got)
		}
	})

	t.Run("Expired entries", func(t *testing.T) {
		err := cache.Set(ctx, "key2", "value2", -1*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		got, err := cache.Get(ctx, "key2")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("Expected empty string for expired entry, got %s", got)
		}
	})

	t.Run("Respect TTL", func(t *testing.T) {
		err := cache.Set(ctx, "key3", "value3", 100*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		got, err := cache.Get(ctx, "key3")
		if err != nil {
			t.Fatal(err)
		}
		if got != "value3" {
			t.Errorf("Expected value3, got %s", got)
		}

		time.Sleep(150 * time.Millisecond)
		got, err = cache.Get(ctx, "key3")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("Expected empty string after TTL, got %s", got)
		}
	})
}
