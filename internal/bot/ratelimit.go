package bot

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	perKey map[int64][]time.Time
}

// newRateLimiter allows up to limit events per window per key.
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &rateLimiter{
		window: window,
		limit:  limit,
		perKey: make(map[int64][]time.Time),
	}
}

// Allow returns whether the key can perform another action now,
// and the time to wait before retrying when denied.
func (l *rateLimiter) Allow(key int64) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Keep only timestamps inside the current window.
	cutoff := now.Add(-l.window)
	recent := pruneOld(l.perKey[key], cutoff)

	// Deny if the key already hit the limit in this window.
	if len(recent) >= l.limit {
		retryAfter := l.window - now.Sub(recent[0])
		if retryAfter < 0 {
			retryAfter = 0
		}
		l.perKey[key] = recent
		return false, retryAfter
	}

	recent = append(recent, now)
	l.perKey[key] = recent
	return true, 0
}

// pruneOld mutates the slice in place to keep only timestamps after cutoff.
func pruneOld(times []time.Time, cutoff time.Time) []time.Time {
	if len(times) == 0 {
		return times
	}
	n := 0
	for _, ts := range times {
		if ts.After(cutoff) {
			times[n] = ts
			n++
		}
	}
	return times[:n]
}
