// Package analytics provides an in-memory evaluation counter for per-flag usage metrics.
// Counters reset on server restart; they are intentionally lightweight (no DB writes
// on the hot evaluation path).
package analytics

import (
	"sync"
	"time"
)

// FlagStats holds evaluation metrics for a single flag in one environment.
type FlagStats struct {
	TotalEvaluations int64            `json:"total_evaluations"`
	UniqueUsers      int64            `json:"unique_users"`
	LastEvaluatedAt  *time.Time       `json:"last_evaluated_at,omitempty"`
	VariationCounts  map[string]int64 `json:"variation_counts"` // variation name/id → count
}

type flagKey struct {
	key         string
	environment string
}

// Counter is a thread-safe, in-memory evaluation counter.
// One instance should be shared across all evaluation handlers.
type Counter struct {
	mu    sync.RWMutex
	stats map[flagKey]*flagStats
}

// internal mutable state per flag — kept separate from the exported FlagStats
// to allow atomic snapshots without holding a lock during JSON serialisation.
type flagStats struct {
	mu               sync.Mutex
	totalEvaluations int64
	users            map[string]struct{} // capped at maxUniqueUsers to bound memory
	lastEvaluatedAt  time.Time
	variationCounts  map[string]int64
}

const maxUniqueUsers = 10_000

// NewCounter creates a new Counter.
func NewCounter() *Counter {
	return &Counter{
		stats: make(map[flagKey]*flagStats),
	}
}

// Record increments the evaluation counter for flagKey/environment.
// userID may be empty (anonymous). variation may be empty for boolean flags.
func (c *Counter) Record(key, environment, userID, variation string) {
	fk := flagKey{key: key, environment: environment}

	c.mu.RLock()
	fs := c.stats[fk]
	c.mu.RUnlock()

	if fs == nil {
		c.mu.Lock()
		// Double-check after acquiring write lock
		if c.stats[fk] == nil {
			c.stats[fk] = &flagStats{
				users:           make(map[string]struct{}),
				variationCounts: make(map[string]int64),
			}
		}
		fs = c.stats[fk]
		c.mu.Unlock()
	}

	now := time.Now()
	fs.mu.Lock()
	fs.totalEvaluations++
	fs.lastEvaluatedAt = now
	if userID != "" && len(fs.users) < maxUniqueUsers {
		fs.users[userID] = struct{}{}
	}
	if variation != "" {
		fs.variationCounts[variation]++
	}
	fs.mu.Unlock()
}

// Snapshot returns a copy of stats for all flags in the given environment.
// If environment is empty, all environments are returned.
func (c *Counter) Snapshot(environment string) map[string]FlagStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]FlagStats)
	for fk, fs := range c.stats {
		if environment != "" && fk.environment != environment {
			continue
		}
		fs.mu.Lock()
		t := fs.lastEvaluatedAt
		vc := make(map[string]int64, len(fs.variationCounts))
		for k, v := range fs.variationCounts {
			vc[k] = v
		}
		stat := FlagStats{
			TotalEvaluations: fs.totalEvaluations,
			UniqueUsers:      int64(len(fs.users)),
			VariationCounts:  vc,
		}
		if !t.IsZero() {
			stat.LastEvaluatedAt = &t
		}
		fs.mu.Unlock()
		result[fk.key] = stat
	}
	return result
}

// SnapshotOne returns stats for a single flag key in an environment.
func (c *Counter) SnapshotOne(key, environment string) (FlagStats, bool) {
	fk := flagKey{key: key, environment: environment}
	c.mu.RLock()
	fs := c.stats[fk]
	c.mu.RUnlock()
	if fs == nil {
		return FlagStats{VariationCounts: map[string]int64{}}, false
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	t := fs.lastEvaluatedAt
	vc := make(map[string]int64, len(fs.variationCounts))
	for k, v := range fs.variationCounts {
		vc[k] = v
	}
	stat := FlagStats{
		TotalEvaluations: fs.totalEvaluations,
		UniqueUsers:      int64(len(fs.users)),
		VariationCounts:  vc,
	}
	if !t.IsZero() {
		stat.LastEvaluatedAt = &t
	}
	return stat, true
}
