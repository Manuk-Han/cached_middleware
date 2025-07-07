package config

type Config struct {
	Cache        CacheConfig        `mapstructure:"cache"`
	EventBroker  EventBrokerConfig  `mapstructure:"event_broker"`
	Invalidation InvalidationConfig `mapstructure:"invalidation"`
}

// Cache
type CacheConfig struct {
	Type  string      `mapstructure:"type"` // redis, memcached
	Redis RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Address    string `mapstructure:"address"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	TTLSeconds int    `mapstructure:"ttl_seconds"`
}

// Event Broker
type EventBrokerConfig struct {
	Type  string      `mapstructure:"type"` // kafka, nats
	Kafka KafkaConfig `mapstructure:"kafka"`
}

type KafkaConfig struct {
	Brokers []string          `mapstructure:"brokers"`
	Topics  []string          `mapstructure:"topics"`
	GroupID string            `mapstructure:"group_id"`
	Reader  KafkaReaderConfig `mapstructure:"reader"`
}

type KafkaReaderConfig struct {
	MinBytes      int `mapstructure:"min_bytes"`
	MaxBytes      int `mapstructure:"max_bytes"`
	MaxWaitMs     int `mapstructure:"max_wait_ms"` // milliseconds
	QueueCapacity int `mapstructure:"queue_capacity"`
}

// Invalidation
type InvalidationConfig struct {
	Strategy  string            `mapstructure:"strategy"` // versioned-key, ttl-aware
	Versioned VersionedStrategy `mapstructure:"versioned"`
}

type VersionedStrategy struct {
	Delimiter      string `mapstructure:"delimiter"`
	DefaultVersion int    `mapstructure:"default_version"`
}
