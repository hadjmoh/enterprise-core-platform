package messaging

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

type AMQPConsumer struct {
	url       string
	queueName string
	pipeline  *pipeline.IngestionPipeline
	logger    *logger.Logger
	conn      *amqp.Connection
	channel   *amqp.Channel
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewAMQPConsumer(url, queueName string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *AMQPConsumer {
	return &AMQPConsumer{
		url:       url,
		queueName: queueName,
		pipeline:  pipe,
		logger:    logger,
	}
}

func (a *AMQPConsumer) Start() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())

	conn, err := amqp.Dial(a.url)
	if err != nil {
		return fmt.Errorf("AMQP connection failed: %w", err)
	}
	a.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("AMQP channel failed: %w", err)
	}
	a.channel = ch

	// Declare queue (idempotent)
	_, err = ch.QueueDeclare(
		a.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("AMQP queue declare failed: %w", err)
	}

	a.logger.Info("AMQP Consumer started", "queue", a.queueName)

	go a.consume()
	return nil
}

func (a *AMQPConsumer) consume() {
	msgs, err := a.channel.Consume(
		a.queueName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		a.logger.Error("AMQP consume failed", err)
		return
	}

	for {
		select {
		case <-a.ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			a.logger.Info("Received AMQP message",
				"exchange", msg.Exchange,
				"routing_key", msg.RoutingKey,
				"size", len(msg.Body),
			)
			
			// Send to ingestion pipeline
			a.pipeline.Process(buffer.Event{
				Timestamp: time.Now().Format(time.RFC3339),
				Source:    fmt.Sprintf("amqp:%s", a.queueName),
				Data: map[string]interface{}{
					"exchange":    msg.Exchange,
					"routing_key": msg.RoutingKey,
					"payload":     string(msg.Body),
					"priority":    msg.Priority,
				},
			})
		}
	}
}

func (a *AMQPConsumer) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.channel != nil {
		a.channel.Close()
	}
	if a.conn != nil {
		a.conn.Close()
		a.logger.Info("AMQP Consumer stopped")
	}
}

func (a *AMQPConsumer) Protocol() string {
	return "AMQP"
}
