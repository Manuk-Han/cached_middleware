package core

import (
	"cache/interface"
)

type CacheService struct {
	cache    _interface.ICacheAdapter
	strategy _interface.IInvalidationStrategy
}

func NewCacheService(c _interface.ICacheAdapter, s _interface.IInvalidationStrategy) *CacheService {
	return &CacheService{cache: c, strategy: s}
}

func (cs *CacheService) Get(topic string, key string) (string, error) {
	actualKey := cs.strategy.GenerateKey(topic, key)
	return cs.cache.Get(actualKey)
}

func (cs *CacheService) Set(topic string, key string, val string, ttl int) error {
	actualKey := cs.strategy.GenerateKey(topic, key)
	ttl = cs.strategy.ComputeTTL(ttl)
	return cs.cache.Set(actualKey, val, ttl)
}

func (cs *CacheService) Invalidate(topic string, key string) error {
	actualKey := cs.strategy.GenerateKey(topic, key)
	return cs.cache.Invalidate(actualKey)
}
