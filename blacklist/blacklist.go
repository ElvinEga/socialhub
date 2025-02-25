package blacklist

import (
	"sync"
	"time"
)

type TokenInfo struct {
	Expiration time.Time
}

var (
	tokens = make(map[string]TokenInfo)
	mutex  sync.RWMutex
)

// Add adds a token to the blacklist with its expiration time.
func Add(token string, expiration time.Time) {
	mutex.Lock()
	defer mutex.Unlock()
	tokens[token] = TokenInfo{Expiration: expiration}
}

// IsBlacklisted checks if a token is in the blacklist.
// It also cleans up expired tokens.
func IsBlacklisted(token string) bool {
	mutex.RLock()
	info, exists := tokens[token]
	mutex.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(info.Expiration) {
		// Remove expired token.
		mutex.Lock()
		delete(tokens, token)
		mutex.Unlock()
		return false
	}
	return true
}
