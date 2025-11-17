package broker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/donnigundala/dgcore/ctxutil"
)

const (
	// defaultKafkaTimeout is the default timeout for operations like polling.
	defaultKafkaTimeout = 100 * time.Millisecond
	// traceIDHeader is the key used for the trace/request ID in Kafka message headers.
	traceIDHeader = "trace_id"
)

// kafkaProvider implements the Provider and KafkaProvider interfaces.
type kafkaProvider struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	config   *Config
	logger   *slog.Logger
}

// newKafkaProvider creates and initializes a new Kafka provider.
func newKafkaProvider(cfg *Config, logger *slog.Logger) (Provider, error) {
	// --- Producer Initialization ---
	producerConfig := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Kafka.Brokers, ","),
		// "acks": "all" is a good default for reliability.
		"acks": "all",
	}
	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		logger.Error("Failed to create Kafka producer", "error", err)
		return nil, fmt.Errorf("kafka: failed to create producer: %w", err)
	}
	logger.Info("Kafka producer created successfully")

	// --- Consumer Initialization ---
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Kafka.Brokers, ","),
		"group.id":          cfg.Kafka.GroupID,
		// Start reading from the beginning of the topic if no offset is stored.
		"auto.offset.reset": "earliest",
		// We will commit offsets manually after successful processing.
		"enable.auto.commit": "false",
	}
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		// Clean up the producer if consumer creation fails.
		producer.Close()
		logger.Error("Failed to create Kafka consumer", "error", err)
		return nil, fmt.Errorf("kafka: failed to create consumer: %w", err)
	}
	logger.Info("Kafka consumer created successfully", "group_id", cfg.Kafka.GroupID)

	return &kafkaProvider{
		producer: producer,
		consumer: consumer,
		config:   cfg,
		logger:   logger,
	}, nil
}

// Publish sends a message to a Kafka topic.
// It injects the trace ID from the context into the message headers.
func (p *kafkaProvider) Publish(ctx context.Context, topic string, msg Message) error {
	log := ctxutil.LoggerFromContext(ctx)

	// Prepare Kafka headers, including the trace ID.
	var headers []kafka.Header
	if traceID := ctxutil.RequestIDFromContext(ctx); traceID != "" {
		headers = append(headers, kafka.Header{Key: traceIDHeader, Value: []byte(traceID)})
	}
	for k, v := range msg.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msg.Body,
		Headers:        headers,
		Timestamp:      msg.Timestamp,
	}

	// Use a delivery channel to make the publish operation synchronous and get immediate feedback.
	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	if err := p.producer.Produce(kafkaMsg, deliveryChan); err != nil {
		log.Error("Failed to produce Kafka message", "topic", topic, "error", err)
		return fmt.Errorf("kafka: failed to produce message: %w", err)
	}

	// Wait for the delivery report.
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Error("Kafka message delivery failed", "topic", topic, "error", m.TopicPartition.Error)
		return fmt.Errorf("kafka: delivery failed: %w", m.TopicPartition.Error)
	}

	log.Debug("Message delivered successfully to Kafka", "topic", *m.TopicPartition.Topic, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset)
	return nil
}

// Subscribe listens for messages on a topic and calls the handler.
// This is a blocking operation.
func (p *kafkaProvider) Subscribe(ctx context.Context, topic string, handler HandlerFunc) error {
	if err := p.consumer.Subscribe(topic, nil); err != nil {
		p.logger.Error("Failed to subscribe to Kafka topic", "topic", topic, "error", err)
		return fmt.Errorf("kafka: failed to subscribe: %w", err)
	}
	p.logger.Info("Subscribed to Kafka topic, waiting for messages...", "topic", topic)

	for {
		select {
		case <-ctx.Done():
			// The context was canceled, so we gracefully shut down the subscription loop.
			p.logger.Info("Context canceled, stopping Kafka subscription loop.", "topic", topic)
			return ctx.Err()
		default:
			// Poll for a message.
			ev := p.consumer.Poll(int(defaultKafkaTimeout.Milliseconds()))
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				p.processMessage(ctx, e, handler)
			case kafka.Error:
				// Errors from the consumer are generally informational.
				p.logger.Error("Received Kafka consumer error", "error", e)
			default:
				p.logger.Debug("Ignored Kafka event", "event", e)
			}
		}
	}
}

// processMessage handles a single received Kafka message.
func (p *kafkaProvider) processMessage(ctx context.Context, kafkaMsg *kafka.Message, handler HandlerFunc) {
	// Extract trace ID from headers to create a new context for this specific message.
	var traceID string
	for _, h := range kafkaMsg.Headers {
		if h.Key == traceIDHeader {
			traceID = string(h.Value)
			break
		}
	}

	// Create a new context and logger for this message processing task.
	handlerCtx := ctxutil.WithRequestID(ctx, traceID)
	log := p.logger.With("request_id", traceID)
	handlerCtx = ctxutil.WithLogger(handlerCtx, log)

	log.Debug("Processing Kafka message", "topic", *kafkaMsg.TopicPartition.Topic)

	// Convert to our standard Message format.
	brokerMsg := Message{
		ID:        string(kafkaMsg.Key), // Assuming key is used as ID
		Body:      kafkaMsg.Value,
		Timestamp: kafkaMsg.Timestamp,
		Headers:   make(map[string]string),
	}
	for _, h := range kafkaMsg.Headers {
		brokerMsg.Headers[h.Key] = string(h.Value)
	}

	// Execute the handler.
	if err := handler(handlerCtx, brokerMsg); err != nil {
		log.Error("Handler failed to process message", "error", err)
		// We do NOT commit the offset here, so the message will be re-processed later.
		return
	}

	// If the handler was successful, commit the offset.
	if _, err := p.consumer.CommitMessage(kafkaMsg); err != nil {
		log.Error("Failed to commit Kafka offset", "error", err)
	}
}

// Ping checks connectivity to the Kafka cluster.
func (p *kafkaProvider) Ping(ctx context.Context) error {
	log := ctxutil.LoggerFromContext(ctx)
	log.Debug("Pinging Kafka cluster")

	// Get metadata for the cluster. A timeout is important here.
	_, err := p.producer.GetMetadata(nil, false, int(5*time.Second.Milliseconds()))
	return err
}

// Close gracefully closes the producer and consumer.
func (p *kafkaProvider) Close() error {
	p.logger.Info("Closing Kafka connections...")
	// The producer's Close flushes any buffered messages.
	p.producer.Close()
	return p.consumer.Close()
}

// Producer provides access to the native Kafka producer.
func (p *kafkaProvider) Producer() *kafka.Producer {
	return p.producer
}

// Consumer provides access to the native Kafka consumer.
func (p *kafkaProvider) Consumer() *kafka.Consumer {
	return p.consumer
}
