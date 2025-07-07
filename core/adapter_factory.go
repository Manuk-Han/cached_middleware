package core

import (
	"cache/config"
	"cache/core/cache_adapter"
	"cache/core/event_broker"
	"cache/core/strategy"
	_interface "cache/interface"
	"errors"
	"fmt"
)

// NewCacheAdapter Cache 어댑터 생성
func NewCacheAdapter(cfg config.CacheConfig) (_interface.ICacheAdapter, error) {
	switch cfg.Type {
	case "redis":
		return cache_adapter.NewRedisAdapter(cfg.Redis), nil
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cfg.Type)
	}
}

// NewEventBroker Event Broker 생성
func NewEventBroker(cfg config.EventBrokerConfig) (_interface.IEventBroker, error) {
	switch cfg.Type {
	case "kafka":
		return event_broker.NewKafkaBroker(cfg.Kafka), nil
	default:
		return nil, fmt.Errorf("unsupported event broker type: %s", cfg.Type)
	}
}

// NewInvalidationStrategy Invalidation 전략 생성
func NewInvalidationStrategy(cfg config.InvalidationConfig) (_interface.IInvalidationStrategy, error) {
	switch cfg.Strategy {
	case "versioned-key":
		return strategy.NewVersionedKeyStrategy(cfg.Versioned), nil
	default:
		return nil, errors.New("unsupported invalidation strategy: " + cfg.Strategy)
	}
}
