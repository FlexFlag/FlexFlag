package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCache(t *testing.T) {
	ttl := time.Minute * 5
	cache := NewMemoryCache(ttl)
	
	assert.NotNil(t, cache)
	assert.Equal(t, ttl, cache.ttl)
	assert.NotNil(t, cache.flags)
	assert.NotNil(t, cache.stopCh)
	
	// Cleanup
	cache.Close()
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Name:        "Test Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}
	
	// Set flag in cache
	cache.Set(ctx, flag.Key, flag.Environment, flag)
	
	// Get flag from cache
	cachedFlag, found := cache.Get(ctx, flag.Key, flag.Environment)
	
	assert.True(t, found)
	assert.NotNil(t, cachedFlag)
	assert.Equal(t, flag.ID, cachedFlag.ID)
	assert.Equal(t, flag.Key, cachedFlag.Key)
	assert.Equal(t, flag.Type, cachedFlag.Type)
	assert.Equal(t, flag.Enabled, cachedFlag.Enabled)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	
	// Try to get non-existent flag
	cachedFlag, found := cache.Get(ctx, "non-existent", "production")
	
	assert.False(t, found)
	assert.Nil(t, cachedFlag)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(time.Millisecond * 50)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "expiring-flag",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
	}
	
	// Set flag in cache
	cache.Set(ctx, flag.Key, flag.Environment, flag)
	
	// Get flag immediately - should be found
	cachedFlag, found := cache.Get(ctx, flag.Key, flag.Environment)
	assert.True(t, found)
	assert.NotNil(t, cachedFlag)
	
	// Wait for expiration
	time.Sleep(time.Millisecond * 100)
	
	// Get flag after expiration - should not be found
	cachedFlag, found = cache.Get(ctx, flag.Key, flag.Environment)
	assert.False(t, found)
	assert.Nil(t, cachedFlag)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "delete-test",
		Environment: "production",
		Type:        types.FlagTypeString,
	}
	
	// Set flag in cache
	cache.Set(ctx, flag.Key, flag.Environment, flag)
	
	// Verify flag is cached
	_, found := cache.Get(ctx, flag.Key, flag.Environment)
	assert.True(t, found)
	
	// Delete flag from cache
	cache.Delete(ctx, flag.Key, flag.Environment)
	
	// Verify flag is no longer cached
	_, found = cache.Get(ctx, flag.Key, flag.Environment)
	assert.False(t, found)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	
	// Add multiple flags
	flags := []*types.Flag{
		{
			ID:          uuid.New().String(),
			Key:         "flag1",
			Environment: "production",
			Type:        types.FlagTypeBoolean,
		},
		{
			ID:          uuid.New().String(),
			Key:         "flag2",
			Environment: "staging",
			Type:        types.FlagTypeString,
		},
		{
			ID:          uuid.New().String(),
			Key:         "flag3",
			Environment: "production",
			Type:        types.FlagTypeNumber,
		},
	}
	
	for _, flag := range flags {
		cache.Set(ctx, flag.Key, flag.Environment, flag)
	}
	
	// Verify all flags are cached
	for _, flag := range flags {
		_, found := cache.Get(ctx, flag.Key, flag.Environment)
		assert.True(t, found)
	}
	
	// Clear cache
	cache.Clear(ctx)
	
	// Verify all flags are removed
	for _, flag := range flags {
		_, found := cache.Get(ctx, flag.Key, flag.Environment)
		assert.False(t, found)
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(time.Minute * 5)
	defer cache.Close()
	
	ctx := context.Background()
	
	// Initially no flags
	stats := cache.Stats()
	assert.Equal(t, 0, stats["cached_flags"])
	assert.Equal(t, float64(300), stats["ttl_seconds"]) // 5 minutes = 300 seconds
	
	// Add some flags
	flag1 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "stats-flag1",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
	}
	flag2 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "stats-flag2",
		Environment: "staging",
		Type:        types.FlagTypeString,
	}
	
	cache.Set(ctx, flag1.Key, flag1.Environment, flag1)
	cache.Set(ctx, flag2.Key, flag2.Environment, flag2)
	
	// Check updated stats
	stats = cache.Stats()
	assert.Equal(t, 2, stats["cached_flags"])
	assert.Equal(t, float64(300), stats["ttl_seconds"])
}

func TestMemoryCache_CacheKey(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	
	// Test that different environments for same key are cached separately
	flag1 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "same-key",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
	}
	flag2 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "same-key",
		Environment: "staging",
		Type:        types.FlagTypeBoolean,
		Enabled:     false,
	}
	
	cache.Set(ctx, flag1.Key, flag1.Environment, flag1)
	cache.Set(ctx, flag2.Key, flag2.Environment, flag2)
	
	// Get production flag
	prodFlag, found := cache.Get(ctx, "same-key", "production")
	assert.True(t, found)
	assert.True(t, prodFlag.Enabled)
	
	// Get staging flag
	stagingFlag, found := cache.Get(ctx, "same-key", "staging")
	assert.True(t, found)
	assert.False(t, stagingFlag.Enabled)
	
	// Verify they are different instances
	assert.NotEqual(t, prodFlag.ID, stagingFlag.ID)
}

func TestCachedFlag_Structure(t *testing.T) {
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Environment: "production",
		Type:        types.FlagTypeJSON,
	}
	
	expiresAt := time.Now().Add(time.Hour)
	cachedFlag := &CachedFlag{
		Flag:      flag,
		ExpiresAt: expiresAt,
	}
	
	assert.Equal(t, flag, cachedFlag.Flag)
	assert.Equal(t, expiresAt, cachedFlag.ExpiresAt)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "concurrent-test",
		Environment: "production",
		Type:        types.FlagTypeNumber,
	}
	
	// Test concurrent writes and reads
	done := make(chan bool, 100)
	
	// Start 50 writers
	for i := 0; i < 50; i++ {
		go func(i int) {
			testFlag := *flag
			testFlag.Key = flag.Key + "-" + string(rune(i))
			cache.Set(ctx, testFlag.Key, testFlag.Environment, &testFlag)
			done <- true
		}(i)
	}
	
	// Start 50 readers
	for i := 0; i < 50; i++ {
		go func(i int) {
			testKey := flag.Key + "-" + string(rune(i))
			cache.Get(ctx, testKey, flag.Environment)
			done <- true
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < 100; i++ {
		<-done
	}
	
	// No assertions needed, just testing for race conditions
}

func TestMemoryCache_CleanupGoroutine(t *testing.T) {
	// This test verifies the cleanup goroutine removes expired entries
	cache := NewMemoryCache(time.Millisecond * 10)
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "cleanup-test",
		Environment: "production",
		Type:        types.FlagTypeVariant,
	}
	
	// Set a flag that will expire quickly
	cache.Set(ctx, flag.Key, flag.Environment, flag)
	
	// Verify flag is initially cached
	_, found := cache.Get(ctx, flag.Key, flag.Environment)
	assert.True(t, found)
	
	// Wait for expiration and cleanup cycle
	time.Sleep(time.Millisecond * 100)
	
	// Manually trigger cleanup check by trying to get the expired flag
	_, found = cache.Get(ctx, flag.Key, flag.Environment)
	assert.False(t, found)
	
	// Close cache to stop cleanup goroutine
	cache.Close()
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "bench-flag",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, flag.Key, flag.Environment, flag)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "bench-flag",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
	}
	
	// Pre-populate cache
	cache.Set(ctx, flag.Key, flag.Environment, flag)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, flag.Key, flag.Environment)
	}
}

func BenchmarkMemoryCache_SetGet(b *testing.B) {
	cache := NewMemoryCache(time.Minute)
	defer cache.Close()
	
	ctx := context.Background()
	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "bench-flag",
		Environment: "production",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, flag.Key, flag.Environment, flag)
		cache.Get(ctx, flag.Key, flag.Environment)
	}
}