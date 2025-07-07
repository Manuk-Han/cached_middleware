package event_broker

import (
	"cache/config"
	"cache/infrautil"
	_interface "cache/interface"
	"cache/logger"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type kafkaBroker struct {
	writers map[string]*kafka.Writer
	reader  *kafka.Reader
	log     *zap.SugaredLogger
	brokers []string
	lock    sync.RWMutex
}

func NewKafkaBroker(cfg config.KafkaConfig) _interface.IEventBroker {
	log := logger.Logger

	writers := make(map[string]*kafka.Writer)
	for _, t := range cfg.Topics {
		createTopicIfNotExists(cfg.Brokers[0], t, log)
		writers[t] = kafkaWriter(cfg.Brokers, t)
		log.Infof("🪄 Initialized writer for topic [%s] from config", t)
	}

	suppress := infrautil.NewSuppressLogger()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       cfg.Brokers,
		GroupID:       cfg.GroupID,
		GroupTopics:   cfg.Topics,
		MinBytes:      cfg.Reader.MinBytes,
		MaxBytes:      cfg.Reader.MaxBytes,
		MaxWait:       time.Duration(cfg.Reader.MaxWaitMs) * time.Millisecond,
		QueueCapacity: cfg.Reader.QueueCapacity,
		ErrorLogger:   kafka.LoggerFunc(func(string, ...interface{}) {}),
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			if !suppress.ShouldLog(msg, 10*time.Second) {
				return
			}
			log.Debugf("[Kafka] "+msg, args...)
		}),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try reading one message just to confirm connectivity, but do not enforce it harshly.
	_, err := reader.FetchMessage(ctx)
	infrautil.LogConnectionResult(log, "Kafka", err)
	if err != nil {
		log.Warn("⚠️ Kafka may not be ready, but continuing anyway...")
		// do not os.Exit(1)
	}

	return &kafkaBroker{
		writers: writers,
		reader:  reader,
		log:     log,
		brokers: cfg.Brokers,
	}
}

func kafkaWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func createTopicIfNotExists(broker string, topic string, log *zap.SugaredLogger) {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Warnf("❌ Failed to dial Kafka broker [%s]: %v", broker, err)
		return
	}
	defer func(conn *kafka.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	controller, err := conn.Controller()
	if err != nil {
		log.Warnf("❌ Failed to get Kafka controller: %v", err)
		return
	}

	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		log.Warnf("❌ Failed to dial Kafka controller [%s]: %v", controllerAddr, err)
		return
	}
	defer func(controllerConn *kafka.Conn) {
		err := controllerConn.Close()
		if err != nil {

		}
	}(controllerConn)

	topicConfigs := []kafka.TopicConfig{{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Warnf("❌ Topic creation failed: %v", err)
	} else {
		log.Infof("✅ Kafka topic [%s] is ready", topic)
	}
}

func (k *kafkaBroker) Publish(topic string, key string) error {
	return k.PublishTo(k.defaultTopic(), key)
}

func (k *kafkaBroker) defaultTopic() string {
	k.lock.RLock()
	for t := range k.writers {
		k.lock.RUnlock()
		return t
	}
	k.lock.RUnlock()
	return "default"
}

func (k *kafkaBroker) PublishTo(topic string, key string) error {
	k.lock.RLock()
	writer, ok := k.writers[topic]
	k.lock.RUnlock()

	if !ok {
		k.lock.Lock()
		writer = kafkaWriter(k.brokers, topic)
		k.writers[topic] = writer
		k.lock.Unlock()
		k.log.Infof("🪄 Created new writer for topic [%s]", topic)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(key),
	}
	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		k.log.Errorf("🔥 Kafka publish error [topic=%s, key=%s]: %v", topic, key, err)
	} else {
		k.log.Infof("📤 Kafka message sent [topic=%s, key=%s]", topic, key)
	}
	return err
}

func (k *kafkaBroker) Subscribe(handler func(topic string, key string)) error {
	ctx := context.Background()
	go infrautil.RunMessageLoop(ctx, k.log, 5, func() (string, error) {
		m, err := k.reader.ReadMessage(ctx)
		return string(m.Value), err
	}, func(msg string) {
		topic := k.reader.Config().Topic
		handler(topic, msg)
	})
	return nil
}

func (k *kafkaBroker) Close() error {
	k.log.Infof("🛑 Kafka broker closing")
	if err := k.reader.Close(); err != nil {
		k.log.Errorf("Kafka reader close error: %v", err)
		return err
	}
	for topic, writer := range k.writers {
		if err := writer.Close(); err != nil {
			k.log.Errorf("Kafka writer close error for topic [%s]: %v", topic, err)
			return err
		}
	}
	k.log.Infof("✅ Kafka broker closed")
	return nil
}
