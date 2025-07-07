package infrautil

import (
	"sync"
	"time"
)

type SuppressLogger struct {
	mu      sync.Mutex
	lastMsg string
	lastAt  time.Time
}

func NewSuppressLogger() *SuppressLogger {
	return &SuppressLogger{}
}

func (s *SuppressLogger) ShouldLog(msg string, interval time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if msg == s.lastMsg && now.Sub(s.lastAt) < interval {
		return false
	}
	s.lastMsg = msg
	s.lastAt = now
	return true
}
