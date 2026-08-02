package services

import (
	"sync"
	"time"
)

type UploadRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string][]time.Time
}

func NewUploadRateLimiter(limit int, window time.Duration) *UploadRateLimiter {
	return &UploadRateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string][]time.Time),
	}
}

func (rl *UploadRateLimiter) Allow(key string, now time.Time) bool {
	if key == "" {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-rl.window)
	timestamps := rl.entries[key]
	kept := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}

	if len(kept) >= rl.limit {
		rl.entries[key] = kept
		return false
	}

	rl.entries[key] = append(kept, now)
	return true
}
