package importer

import "sync"

// The registry is intentionally a flat, in-process slice. Importers are
// expected to register themselves at init() time from their subpackage; the
// HTTP layer only ever reads the list, never mutates it after startup. A
// mutex guards Register so concurrent init() across packages stays safe.
var (
	mu        sync.RWMutex
	importers []Importer
)

// Register adds `i` to the registry. Call this from a subpackage's init()
// function (e.g. internal/importer/alipay/alipay.go) to make it available to
// the HTTP API and the preview UI. The stub importer used by framework tests
// is the deliberate exception: it lives behind RegisterForTesting so it
// never becomes part of a production build.
//
// Re-registration is a no-op rather than a panic: tests sometimes pull in
// the same importer package twice through different routes; we don't want
// that to crash the binary. The first registration wins.
func Register(i Importer) {
	mu.Lock()
	defer mu.Unlock()
	for _, existing := range importers {
		if existing.Code() == i.Code() {
			return
		}
	}
	importers = append(importers, i)
}

// All returns a copy of every registered importer in registration order.
// Callers may safely iterate without holding the registry lock.
func All() []Importer {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Importer, len(importers))
	copy(out, importers)
	return out
}

// Detect returns every importer whose Detect() returned true for the given
// file. Multiple matches are possible (e.g. several Chinese-bank CSVs share
// a similar header); the UI lets the user pick. An empty slice means "no
// importer recognised this file — fall back to Handlebars templates".
func Detect(filename string, content []byte) []Importer {
	mu.RLock()
	defer mu.RUnlock()
	var matched []Importer
	for _, i := range importers {
		if i.Detect(filename, content) {
			matched = append(matched, i)
		}
	}
	return matched
}

// ByCode returns the importer with the given code, or nil if none match.
// Used by the /api/import/parse handler to dispatch by user-supplied code.
func ByCode(code string) Importer {
	mu.RLock()
	defer mu.RUnlock()
	for _, i := range importers {
		if i.Code() == code {
			return i
		}
	}
	return nil
}

// ResetForTesting clears the registry. EXPORTED but for tests only — every
// caller in production code must instead let importers register themselves
// at init() time and leave the registry append-only. The name is verbose on
// purpose so a casual reader spots the smell at the call site.
func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	importers = nil
}
