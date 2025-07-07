package cache_adapter

import (
	"cache/config"
	_interface "cache/interface"
	"cache/logger"
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type redisAdapter struct {
	client  *redis.Client
	ttl     int
	log     *zap.SugaredLogger
	logOnce sync.Once
}

func NewRedisAdapter(cfg config.RedisConfig) _interface.ICacheAdapter {
	log := logger.Logger

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	failCount := 0
	maxFails := 3

	for i := 1; i <= maxFails; i++ {
		log.Infof("🔄 Attempting to connect to Redis %d/%d...", i, maxFails)
		err := rdb.Ping(ctx).Err()
		if err == nil {
			log.Infof("✅ Redis connected")
			return &redisAdapter{
				client: rdb,
				ttl:    cfg.TTLSeconds,
				log:    log,
			}
		}
		failCount++
		time.Sleep(2 * time.Second)
	}

	os.Exit(1)
	return nil
}

func (r *redisAdapter) Get(key string) (string, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		r.log.Infof("🔍 Cache miss [key=%s]", key)
		return "", nil
	}
	if err != nil {
		r.log.Errorf("❗ Redis GET error [key=%s]: %v", key, err)
		return "", err
	}
	r.log.Infof("✅ Cache hit [key=%s]", key)
	return val, nil
}

func (r *redisAdapter) Set(key string, value string, ttlSeconds int) error {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		r.log.Errorf("❗ Redis SET error [key=%s]: %v", key, err)
		return err
	}
	r.log.Infof("📌 Cache set [key=%s, ttl=%ds]", key, ttlSeconds)
	return nil
}

func (r *redisAdapter) Invalidate(key string) error {
	ctx := context.Background()
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.log.Errorf("❗ Redis DEL error [key=%s]: %v", key, err)
		return err
	}
	r.log.Infof("🚫 Cache invalidated [key=%s]", key)
	return nil
}
