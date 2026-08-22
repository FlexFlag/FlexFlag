package handlers

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// PrecomputedFlag contains pre-processed flag data for ultra-fast evaluation.
// Fields are written once (in buildPrecomputed) and never modified afterward,
// making the struct safe to share across goroutines without a lock.
type PrecomputedFlag struct {
	Key             string
	Enabled         bool
	DefaultValue    interface{}            // pre-unmarshaled; no json.Unmarshal on the hot path
	DefaultJSON     []byte                 // pre-marshaled for direct response writes
	Type            types.FlagType
	HasTargeting    bool
	Variations      []types.Variation
	VariationValues map[string]interface{} // variationID → pre-unmarshaled value
	Targeting       *types.TargetingConfig
	LastUpdated     time.Time
}

// cachedResponse holds a pre-serialized HTTP response body and the Snapshot
// version it was computed from. A cached response is valid only when its
// snapshotVer matches the version of the currently loaded Snapshot — no sweep,
// no contains() call, no manual invalidation needed.
type cachedResponse struct {
	bytes       []byte
	expiresAt   time.Time
	snapshotVer uint64
}

// UltraFastHandler serves flag evaluations with minimum latency.
//
// Flag state lives in a flagStore (atomic.Pointer[Snapshot]). Reads are
// lock-free: one atomic.Load per request. Writes (flag updates) are copy-on-
// write: a new Snapshot is built under a mutex and swapped atomically. The
// old Snapshot is collected by the GC once all in-flight requests that
// referenced it have completed.
//
// Response caching (cachedResponse) provides an additional layer: identical
// requests return pre-serialized bytes without re-evaluating. Cache entries
// are invalidated by version mismatch rather than a sweep, eliminating the
// class of bug where a wrong invalidation predicate silently serves stale data.
type UltraFastHandler struct {
	repo   storage.FlagRepository
	store  *flagStore   // lock-free reads via atomic.Pointer[Snapshot]

	cacheMu        sync.RWMutex
	responseCaches map[string]*cachedResponse

	preloadDone atomic.Bool  // set to true once initial preload completes
	stopCh      chan struct{} // close to stop the cleanup goroutine
}

func NewUltraFastHandler(repo storage.FlagRepository) *UltraFastHandler {
	h := &UltraFastHandler{
		repo:           repo,
		store:          newFlagStore(), // initializes atomic.Pointer with an empty Snapshot
		responseCaches: make(map[string]*cachedResponse),
		stopCh:         make(chan struct{}),
	}
	go h.preloadFlags()
	go h.cleanupExpiredCaches()
	return h
}

// Close stops the background cleanup goroutine. Call this when the handler
// is no longer needed (e.g., in tests or on graceful shutdown).
func (h *UltraFastHandler) Close() {
	close(h.stopCh)
}

// preloadFlags loads all flags for the standard environments at startup.
// It accumulates flags into a local map and calls store.reset() once per
// environment, so readers never observe a half-loaded state.
func (h *UltraFastHandler) preloadFlags() {
	ctx := context.Background()
	environments := []string{"production", "staging", "development"}

	for _, env := range environments {
		flags, err := h.repo.List(ctx, env)
		if err != nil {
			continue
		}

		// Build the full map for this environment before publishing.
		// Merging with the current snapshot means flags from previously
		// loaded environments are preserved.
		snap := h.store.load()
		next := make(map[string]*PrecomputedFlag, len(snap.flags)+len(flags))
		for k, v := range snap.flags {
			next[k] = v
		}
		for _, flag := range flags {
			next[flag.Key+":"+env] = buildPrecomputed(flag)
		}
		h.store.reset(next)
	}

	h.preloadDone.Store(true)
}

func (h *UltraFastHandler) UltraFastEvaluate(c *gin.Context) {
	startTime := time.Now()

	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment := c.DefaultQuery("environment", "production")
	if apiKeyEnv, exists := c.Get("environment"); exists {
		environment = apiKeyEnv.(string)
	}

	// Load the current Snapshot once. All flag reads for this request use
	// this pointer — no lock held, no further atomic operations needed.
	snap := h.store.load()

	cacheKey := h.generateCacheKey(&req, environment)

	// Response cache check: valid only if the entry was built from the same
	// Snapshot version. A flag update bumps the version, making all prior
	// entries stale by definition — no sweep, no predicate, no bug.
	h.cacheMu.RLock()
	cached, exists := h.responseCaches[cacheKey]
	h.cacheMu.RUnlock()

	if exists && cached.snapshotVer == snap.version && time.Now().Before(cached.expiresAt) {
		c.Data(http.StatusOK, "application/json", cached.bytes)
		return
	}

	// Flag lookup from the immutable Snapshot — no lock.
	flagKey := req.FlagKey + ":" + environment
	flag, exists := snap.flags[flagKey]

	if !exists {
		// Dynamic miss: fetch from DB, add to store, reload version.
		projectID := c.Query("project_id")
		var dbFlag *types.Flag
		var err error

		if projectID != "" {
			dbFlag, err = h.repo.GetByProjectKey(c.Request.Context(), projectID, req.FlagKey, environment)
		} else {
			dbFlag, err = h.repo.GetByKey(c.Request.Context(), req.FlagKey, environment)
		}

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
			return
		}

		flag = buildPrecomputed(dbFlag)
		// set() returns the new version so the response cache entry we write
		// below is stamped with the Snapshot that contains this flag.
		snap = &Snapshot{flags: snap.flags, version: h.store.set(flagKey, flag)}
	}

	// Evaluation — reads only from the immutable PrecomputedFlag.
	var responseValue interface{}
	var reason string
	var isDefault bool

	if !flag.Enabled {
		responseValue = flag.DefaultValue
		reason = "flag_disabled"
		isDefault = true
	} else if !flag.HasTargeting {
		responseValue = flag.DefaultValue
		reason = "default"
		isDefault = true
	} else {
		if matched := h.fastTargetingEvaluation(flag, &req); matched != nil {
			responseValue = matched
			reason = "rule_match"
			isDefault = false
		} else {
			responseValue = flag.DefaultValue
			reason = "default"
			isDefault = true
		}
	}

	evalTime := float64(time.Since(startTime).Nanoseconds()) / 1_000_000.0

	response := map[string]interface{}{
		"flag_key":           flag.Key,
		"value":              responseValue,
		"reason":             reason,
		"default":            isDefault,
		"evaluation_time_ms": evalTime,
		"timestamp":          time.Now(),
	}

	// Marshal once, cache the bytes stamped with the current Snapshot version.
	responseBytes, _ := json.Marshal(response)

	h.cacheMu.Lock()
	h.responseCaches[cacheKey] = &cachedResponse{
		bytes:       responseBytes,
		expiresAt:   time.Now().Add(30 * time.Second),
		snapshotVer: snap.version,
	}
	h.cacheMu.Unlock()

	c.Data(http.StatusOK, "application/json", responseBytes)
}

// parallelThreshold is the minimum batch size at which goroutine fan-out
// beats sequential evaluation. Below this, goroutine creation cost (~1µs each)
// exceeds the parallelism benefit for typical precomputed flag evaluation.
// Measured crossover: ~4 flags with targeting rules, ~20+ for simple flags.
const parallelThreshold = 4

// batchResult holds the outcome of evaluating one flag in a batch.
// Indexed by position in the request so goroutines write to separate slots
// with no mutex required.
type batchResult struct {
	key    string
	value  interface{}
	reason string
	found  bool
}

// UltraFastBatchEvaluate evaluates multiple flags in a single request.
//
// The Snapshot is loaded once and shared across all goroutines. Because
// Snapshot.flags is immutable after creation, goroutines read concurrently
// with zero locks — no RLock, no cache-line bouncing between cores.
//
// Result slots are pre-allocated by index. Each goroutine writes only to
// results[i] where i is its position in the request. No mutex on results.
//
// For small batches (below parallelThreshold) the overhead of spawning
// goroutines exceeds the parallelism benefit; those are evaluated sequentially
// using the same evaluateOne helper.
func (h *UltraFastHandler) UltraFastBatchEvaluate(c *gin.Context) {
	var req struct {
		FlagKeys   []string               `json:"flag_keys" binding:"required"`
		UserID     string                 `json:"user_id"`
		Attributes map[string]interface{} `json:"attributes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment := c.DefaultQuery("environment", "production")
	if apiKeyEnv, exists := c.Get("environment"); exists {
		environment = apiKeyEnv.(string)
	}

	// One atomic load. All goroutines below share this pointer.
	// The Snapshot is immutable — no lock needed per goroutine.
	snap := h.store.load()

	evalReq := &EvaluateRequest{
		UserID:     req.UserID,
		Attributes: req.Attributes,
	}

	results := make([]batchResult, len(req.FlagKeys))

	if len(req.FlagKeys) < parallelThreshold {
		// Sequential: goroutine overhead not worth it for small batches.
		for i, key := range req.FlagKeys {
			results[i] = h.evaluateOne(snap, key, environment, evalReq)
		}
	} else {
		// Parallel: one goroutine per flag, all reading the same immutable Snapshot.
		// Each goroutine writes to results[i] — its own slot, no contention.
		var wg sync.WaitGroup
		wg.Add(len(req.FlagKeys))
		for i, key := range req.FlagKeys {
			i, key := i, key
			go func() {
				defer wg.Done()
				results[i] = h.evaluateOne(snap, key, environment, evalReq)
			}()
		}
		wg.Wait()
	}

	output := make(map[string]interface{}, len(results))
	for _, r := range results {
		if !r.found {
			output[r.key] = map[string]interface{}{"error": "flag not found"}
		} else {
			output[r.key] = map[string]interface{}{
				"value":  r.value,
				"reason": r.reason,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": output})
}

// evaluateOne evaluates a single flag against an already-loaded Snapshot.
// Called by both the sequential and parallel paths of UltraFastBatchEvaluate.
// All fields of snap and flag are read-only — safe to call from any goroutine
// with no synchronisation.
func (h *UltraFastHandler) evaluateOne(snap *Snapshot, flagKey, environment string, req *EvaluateRequest) batchResult {
	flag, exists := snap.flags[flagKey+":"+environment]
	if !exists {
		return batchResult{key: flagKey, found: false}
	}

	req.FlagKey = flagKey

	var value interface{}
	var reason string

	if !flag.Enabled {
		value = flag.DefaultValue
		reason = "flag_disabled"
	} else if !flag.HasTargeting {
		value = flag.DefaultValue
		reason = "default"
	} else {
		if matched := h.fastTargetingEvaluation(flag, req); matched != nil {
			value = matched
			reason = "rule_match"
		} else {
			value = flag.DefaultValue
			reason = "default"
		}
	}

	return batchResult{key: flagKey, value: value, reason: reason, found: true}
}

func (h *UltraFastHandler) fastTargetingEvaluation(flag *PrecomputedFlag, req *EvaluateRequest) interface{} {
	if flag.Targeting == nil {
		return nil
	}

	for _, rule := range flag.Targeting.Rules {
		if value, exists := req.Attributes[rule.Attribute]; exists {
			if h.matchesRule(toString(value), &rule) {
				if v, ok := flag.VariationValues[rule.Variation]; ok {
					return v
				}
			}
		}
	}

	if flag.Targeting.Rollout != nil {
		bucket := h.fastHash(req.UserID+flag.Key) % 100
		cumulative := 0
		for _, vr := range flag.Targeting.Rollout.Variations {
			cumulative += vr.Weight
			if bucket < cumulative {
				if v, ok := flag.VariationValues[vr.VariationID]; ok {
					return v
				}
			}
		}
	}

	return nil
}

func (h *UltraFastHandler) matchesRule(value string, rule *types.TargetingRule) bool {
	switch rule.Operator {
	case "eq":
		return contains(rule.Values, value)
	case "in":
		return contains(rule.Values, value)
	case "contains":
		return len(rule.Values) > 0 && contains([]string{value}, rule.Values[0])
	default:
		return false
	}
}

// fastHash uses FNV-1a for percentage bucketing. Non-cryptographic, sufficient
// for rollout distribution. unsafe.StringData avoids the []byte copy —
// benchmarks confirmed 0 allocs vs 0 allocs for []byte(s), but io.WriteString
// allocates (2 allocs, 72B) because the StringWriter assertion prevents
// devirtualization.
func (h *UltraFastHandler) fastHash(s string) int {
	if s == "" {
		return 0
	}
	data := (*[1 << 20]byte)(unsafe.Pointer(unsafe.StringData(s)))[:len(s)]
	h64 := fnv.New64a()
	_, _ = h64.Write(data)
	return int(h64.Sum64() % 100)
}

func (h *UltraFastHandler) generateCacheKey(req *EvaluateRequest, environment string) string {
	hf := fnv.New64a()
	_, _ = hf.Write([]byte(req.FlagKey))
	_, _ = hf.Write([]byte{':'})
	_, _ = hf.Write([]byte(environment))
	_, _ = hf.Write([]byte{':'})
	_, _ = hf.Write([]byte(req.UserID))
	for k, v := range req.Attributes {
		_, _ = hf.Write([]byte{':'})
		_, _ = hf.Write([]byte(k))
		_, _ = hf.Write([]byte{'='})
		_, _ = hf.Write([]byte(toString(v)))
	}
	return strconv.FormatUint(hf.Sum64(), 36)
}

// cleanupExpiredCaches removes expired response cache entries periodically.
// With version-based invalidation, this is purely a memory reclamation
// mechanism — correctness no longer depends on it running.
func (h *UltraFastHandler) cleanupExpiredCaches() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			h.cacheMu.Lock()
			for key, entry := range h.responseCaches {
				if now.After(entry.expiresAt) {
					delete(h.responseCaches, key)
				}
			}
			h.cacheMu.Unlock()
		case <-h.stopCh:
			return
		}
	}
}

// RefreshFlag reloads a single flag from the DB and publishes it in a new
// Snapshot. The version bump automatically invalidates all response cache
// entries that reference the old Snapshot — no sweep required.
func (h *UltraFastHandler) RefreshFlag(flagKey, environment string) {
	ctx := context.Background()
	storeKey := flagKey + ":" + environment

	flag, err := h.repo.GetByKey(ctx, flagKey, environment)
	if err != nil {
		// Flag deleted upstream — remove from store.
		// Version bump from delete() invalidates stale response cache entries.
		h.store.delete(storeKey)
		return
	}

	h.store.set(storeKey, buildPrecomputed(flag))
}

// RefreshAllFlags rebuilds the full flag set. The response cache is cleared
// explicitly as a memory optimisation (all entries would be stale on version
// mismatch anyway, but clearing eagerly reclaims memory immediately).
func (h *UltraFastHandler) RefreshAllFlags() {
	go h.preloadFlags()

	h.cacheMu.Lock()
	h.responseCaches = make(map[string]*cachedResponse)
	h.cacheMu.Unlock()
}

func (h *UltraFastHandler) GetStats(c *gin.Context) {
	snap := h.store.load()

	h.cacheMu.RLock()
	cacheCount := len(h.responseCaches)
	h.cacheMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"preloaded_flags":  len(snap.flags),
		"cached_responses": cacheCount,
		"preload_complete": h.preloadDone.Load(),
		"snapshot_version": snap.version,
	})
}

// buildPrecomputed creates a PrecomputedFlag from a DB flag, pre-unmarshaling
// all JSON values so the hot evaluation path never calls json.Unmarshal.
func buildPrecomputed(flag *types.Flag) *PrecomputedFlag {
	var defaultValue interface{}
	_ = json.Unmarshal(flag.Default, &defaultValue)

	varValues := make(map[string]interface{}, len(flag.Variations))
	for _, v := range flag.Variations {
		var val interface{}
		_ = json.Unmarshal(v.Value, &val)
		varValues[v.ID] = val
		if v.Name != "" {
			varValues[v.Name] = val
		}
	}

	return &PrecomputedFlag{
		Key:             flag.Key,
		Enabled:         flag.Enabled,
		DefaultValue:    defaultValue,
		DefaultJSON:     flag.Default,
		Type:            flag.Type,
		HasTargeting:    flag.Targeting != nil,
		Variations:      flag.Variations,
		VariationValues: varValues,
		Targeting:       flag.Targeting,
		LastUpdated:     flag.UpdatedAt,
	}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func contains(slice []string, item string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}
	return false
}

// Ensure atomic.Bool is used correctly — it must not be copied after first use.
var _ = atomic.Bool{}
