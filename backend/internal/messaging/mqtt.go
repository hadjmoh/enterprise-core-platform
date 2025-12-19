package messaging

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	broker   string
	clientID string
	topic    string
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	client   mqtt.Client
}

func NewMQTTClient(broker, clientID, topic string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *MQTTClient {
	return &MQTTClient{
		broker:   broker,
		clientID: clientID,
		topic:    topic,
		pipeline: pipe,
		logger:   logger,
	}
}

func (m *MQTTClient) Start() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.broker)
	opts.SetClientID(m.clientID)
	opts.SetDefaultPublishHandler(m.messageHandler)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("MQTT connection failed: %w", token.Error())
	}

	m.logger.Info("MQTT Client connected", "broker", m.broker)

	// Subscribe to topic
	if token := m.client.Subscribe(m.topic, 1, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("MQTT subscription failed: %w", token.Error())
	}

	m.logger.Info("MQTT subscribed to topic", "topic", m.topic)
	return nil
}

func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	m.logger.Info("Received MQTT message",
		"topic", msg.Topic(),
		"payload_len", len(msg.Payload()),
	)
	
	// Send to ingestion pipeline
	m.pipeline.Process(buffer.Event{
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    fmt.Sprintf("mqtt:%s", msg.Topic()),
		Data: map[string]interface{}{
			"topic":   msg.Topic(),
			"payload": string(msg.Payload()),
			"qos":     msg.Qos(),
		},
	})
}

func (m *MQTTClient) Stop() {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250)
		m.logger.Info("MQTT Client disconnected")
	}
}

func (m *MQTTClient) Protocol() string {
	return "MQTT"
}
