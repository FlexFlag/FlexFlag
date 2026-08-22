package handlers

import (
	"sync"
	"sync/atomic"
)

// Snapshot is an immutable point-in-time view of all precomputed flags.
//
// Once a Snapshot is stored via flagStore, its fields are never modified.
// Readers load a pointer to the current Snapshot with one atomic instruction
// and hold no lock for the duration of their read. Writers build a new
// Snapshot (copy-on-write) and swap the pointer atomically.
//
// The old Snapshot is garbage-collected once all readers that loaded it have
// finished their requests. This is why the pattern works without manual memory
// management: Go's GC is the reclamation mechanism.
type Snapshot struct {
	flags   map[string]*PrecomputedFlag // flag_key:env → PrecomputedFlag; immutable after creation
	version uint64                      // bumped on every write; used to invalidate response cache entries
}

// flagStore wraps atomic.Pointer[Snapshot] with copy-on-write update semantics.
//
// Read path: load() — one atomic.Pointer.Load(), zero locks, one CPU instruction.
// Write path: set/delete/reset — acquire writeMu, copy the flag map, store a new
// Snapshot, release writeMu. Writers serialize against each other; readers are
// never involved.
type flagStore struct {
	ptr     atomic.Pointer[Snapshot]
	writeMu sync.Mutex // serializes concurrent writers only; readers never acquire this
}

// newFlagStore returns a flagStore with an empty initial Snapshot.
//
// The initial Store must happen before any goroutine calls load(), because
// atomic.Pointer's zero value holds a nil pointer and load() does not nil-check.
func newFlagStore() *flagStore {
	fs := &flagStore{}
	fs.ptr.Store(&Snapshot{
		flags:   make(map[string]*PrecomputedFlag),
		version: 0,
	})
	return fs
}

// load returns the current Snapshot. The caller must not modify the returned
// Snapshot or its flag map. No lock is held after this call returns.
func (fs *flagStore) load() *Snapshot {
	return fs.ptr.Load()
}

// set adds or replaces one flag in a new Snapshot and returns the new version.
//
// The caller uses the returned version to stamp any response cache entry written
// immediately after, binding the cached response to this exact Snapshot.
func (fs *flagStore) set(key string, flag *PrecomputedFlag) uint64 {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	old := fs.ptr.Load()
	next := make(map[string]*PrecomputedFlag, len(old.flags)+1)
	for k, v := range old.flags {
		next[k] = v
	}
	next[key] = flag

	ver := old.version + 1
	fs.ptr.Store(&Snapshot{flags: next, version: ver})
	return ver
}

// delete removes a flag in a new Snapshot and returns the new version.
//
// Readers that loaded the old Snapshot before this call still see the deleted
// flag for the duration of their request. Readers that load after this call
// do not see it. There is no in-between state.
func (fs *flagStore) delete(key string) uint64 {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	old := fs.ptr.Load()
	next := make(map[string]*PrecomputedFlag, len(old.flags))
	for k, v := range old.flags {
		if k != key {
			next[k] = v
		}
	}

	ver := old.version + 1
	fs.ptr.Store(&Snapshot{flags: next, version: ver})
	return ver
}

// reset replaces the entire flag set atomically and returns the new version.
//
// Used by preloadFlags to publish a complete environment's flags in one atomic
// operation. Readers never see a half-loaded state.
func (fs *flagStore) reset(allFlags map[string]*PrecomputedFlag) uint64 {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	old := fs.ptr.Load()
	ver := old.version + 1
	fs.ptr.Store(&Snapshot{flags: allFlags, version: ver})
	return ver
}
