package core

import (
	"cache/interface"
)

type EventListener struct {
	broker _interface.IEventBroker
	cache  *CacheService
}

// NewEventListener 생성자 함수: 의존성 주입
func NewEventListener(b _interface.IEventBroker, c *CacheService) *EventListener {
	return &EventListener{
		broker: b,
		cache:  c,
	}
}

// Start 브로커로부터 메시지를 수신해 invalidate 처리
func (e *EventListener) Start() {
	_ = e.broker.Subscribe(func(topic string, key string) {
		_ = e.cache.Invalidate(topic, key)
	})
}
