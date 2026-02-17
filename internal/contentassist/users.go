package contentassist

import (
	"strings"
	"sync"
)

var (
	mu sync.RWMutex
	// default static cache — can be replaced via SetUserCache in tests or init
	userCache = []string{"alice", "bob", "carol", "dave", "eve", "mallory", "trent"}
)

// SetUserCache replaces the internal user suggestion cache.
func SetUserCache(items []string) {
	mu.Lock()
	defer mu.Unlock()
	userCache = make([]string, len(items))
	copy(userCache, items)
}

// SuggestUsers returns up to 5 suggestions that start with the given prefix (case-insensitive).
func SuggestUsers(prefix string) []string {
	mu.RLock()
	defer mu.RUnlock()
	p := strings.ToLower(strings.TrimSpace(prefix))
	if p == "" {
		// return top 5
		n := 5
		if len(userCache) < n {
			n = len(userCache)
		}
		return append([]string{}, userCache[:n]...)
	}
	out := make([]string, 0, 5)
	for _, u := range userCache {
		if strings.HasPrefix(strings.ToLower(u), p) {
			out = append(out, u)
			if len(out) >= 5 {
				break
			}
		}
	}
	return out
}
