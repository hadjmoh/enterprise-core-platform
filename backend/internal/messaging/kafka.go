package messaging

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	brokers  []string
	topic    string
	groupID  string
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	reader   *kafka.Reader
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewKafkaConsumer(brokers []string, topic, groupID string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *KafkaConsumer {
	return &KafkaConsumer{
		brokers:  brokers,
		topic:    topic,
		groupID:  groupID,
		pipeline: pipe,
		logger:   logger,
	}
}

func (k *KafkaConsumer) Start() error {
	k.ctx, k.cancel = context.WithCancel(context.Background())

	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  k.brokers,
		Topic:    k.topic,
		GroupID:  k.groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	k.logger.Info("Kafka Consumer started", "topic", k.topic, "brokers", k.brokers)

	go k.consume()
	return nil
}

func (k *KafkaConsumer) consume() {
	for {
		msg, err := k.reader.ReadMessage(k.ctx)
		if err != nil {
			select {
			case <-k.ctx.Done():
				return
			default:
				k.logger.Error("Kafka read error", err)
				continue
			}
		}

		k.logger.Info("Received Kafka message",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"size", len(msg.Value),
		)
		
		// Send to ingestion pipeline
		k.pipeline.Process(buffer.Event{
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("kafka:%s", msg.Topic),
			Data: map[string]interface{}{
				"topic":     msg.Topic,
				"partition": msg.Partition,
				"offset":    msg.Offset,
				"payload":   string(msg.Value),
				"key":       string(msg.Key),
			},
		})
	}
}

func (k *KafkaConsumer) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
	if k.reader != nil {
		k.reader.Close()
		k.logger.Info("Kafka Consumer stopped")
	}
}

func (k *KafkaConsumer) Protocol() string {
	return "Kafka"
}
