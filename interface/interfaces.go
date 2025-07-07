package _interface

type ICacheAdapter interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSeconds int) error
	Invalidate(key string) error
}

type IEventBroker interface {
	Publish(topic string, key string) error
	PublishTo(topic string, key string) error
	Subscribe(handler func(topic string, key string)) error
}

type IInvalidationStrategy interface {
	GenerateKey(topic string, key string) string
	ComputeTTL(baseTTL int) int
}
