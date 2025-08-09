package cache

import (
	"context"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

type MemoryCache struct {
	flags    map[string]*CachedFlag
	mu       sync.RWMutex
	ttl      time.Duration
	stopCh   chan struct{}
}

type CachedFlag struct {
	Flag      *types.Flag
	ExpiresAt time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		flags:  make(map[string]*CachedFlag),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

func (c *MemoryCache) Get(ctx context.Context, key, environment string) (*types.Flag, bool) {
	cacheKey := key + ":" + environment
	
	c.mu.RLock()
	cached, exists := c.flags[cacheKey]
	c.mu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	// Check expiration
	if time.Now().After(cached.ExpiresAt) {
		c.mu.Lock()
		delete(c.flags, cacheKey)
		c.mu.Unlock()
		return nil, false
	}
	
	return cached.Flag, true
}

func (c *MemoryCache) Set(ctx context.Context, key, environment string, flag *types.Flag) {
	cacheKey := key + ":" + environment
	
	c.mu.Lock()
	c.flags[cacheKey] = &CachedFlag{
		Flag:      flag,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *MemoryCache) Delete(ctx context.Context, key, environment string) {
	cacheKey := key + ":" + environment
	
	c.mu.Lock()
	delete(c.flags, cacheKey)
	c.mu.Unlock()
}

func (c *MemoryCache) Clear(ctx context.Context) {
	c.mu.Lock()
	c.flags = make(map[string]*CachedFlag)
	c.mu.Unlock()
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, cached := range c.flags {
				if now.After(cached.ExpiresAt) {
					delete(c.flags, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

func (c *MemoryCache) Close() {
	close(c.stopCh)
}

func (c *MemoryCache) Stats() map[string]interface{} {
	c.mu.RLock()
	count := len(c.flags)
	c.mu.RUnlock()
	
	return map[string]interface{}{
		"cached_flags": count,
		"ttl_seconds":  c.ttl.Seconds(),
	}
}