package broker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/donnigundala/dgcore/ctxutil"
	amqp "github.com/rabbitmq/amqp091-go"
)

// rabbitmqProvider implements the Provider and RabbitMQProvider interfaces.
type rabbitmqProvider struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	config *Config
	logger *slog.Logger
}

// newRabbitMQProvider creates and initializes a new RabbitMQ provider.
func newRabbitMQProvider(cfg *Config, logger *slog.Logger) (Provider, error) {
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "url", cfg.RabbitMQ.URL, "error", err)
		return nil, fmt.Errorf("rabbitmq: failed to connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // Clean up connection if channel creation fails
		logger.Error("Failed to open a channel", "error", err)
		return nil, fmt.Errorf("rabbitmq: failed to open channel: %w", err)
	}

	logger.Info("RabbitMQ connection and channel established successfully")

	return &rabbitmqProvider{
		conn:   conn,
		ch:     ch,
		config: cfg,
		logger: logger,
	}, nil
}

// Publish sends a message to a RabbitMQ exchange with a specific routing key.
// The topic parameter is used as the routing key.
func (p *rabbitmqProvider) Publish(ctx context.Context, topic string, msg Message) error {
	log := ctxutil.LoggerFromContext(ctx)

	// Prepare AMQP headers, including the trace ID.
	headers := amqp.Table{}
	if traceID := ctxutil.RequestIDFromContext(ctx); traceID != "" {
		headers[traceIDHeader] = traceID
	}
	for k, v := range msg.Headers {
		headers[k] = v
	}

	amqpMsg := amqp.Publishing{
		Headers:      headers,
		ContentType:  "application/octet-stream",
		Body:         msg.Body,
		Timestamp:    msg.Timestamp,
		MessageId:    msg.ID,
		DeliveryMode: amqp.Persistent, // Good default for reliability
	}

	err := p.ch.PublishWithContext(ctx,
		p.config.RabbitMQ.Exchange, // exchange
		topic,                      // routing key (topic)
		false,                      // mandatory
		false,                      // immediate
		amqpMsg,
	)

	if err != nil {
		log.Error("Failed to publish message to RabbitMQ", "topic", topic, "exchange", p.config.RabbitMQ.Exchange, "error", err)
		return fmt.Errorf("rabbitmq: failed to publish: %w", err)
	}

	log.Debug("Message published successfully to RabbitMQ", "topic", topic, "exchange", p.config.RabbitMQ.Exchange)
	return nil
}

// Subscribe consumes messages from a queue bound to the topic and calls the handler.
// In RabbitMQ, this assumes a queue has been declared and is bound to the exchange with the topic as a binding key.
// This is a blocking operation.
func (p *rabbitmqProvider) Subscribe(ctx context.Context, topic string, handler HandlerFunc) error {
	// For simplicity, we assume the queue name is the same as the topic.
	// In a real-world scenario, this might need to be more configurable.
	q, err := p.ch.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		p.logger.Error("Failed to declare queue for subscription", "queue", topic, "error", err)
		return fmt.Errorf("rabbitmq: failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange with the topic as the routing key.
	err = p.ch.QueueBind(
		q.Name,                     // queue name
		topic,                      // routing key
		p.config.RabbitMQ.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		p.logger.Error("Failed to bind queue", "queue", q.Name, "exchange", p.config.RabbitMQ.Exchange, "error", err)
		return fmt.Errorf("rabbitmq: failed to bind queue: %w", err)
	}

	msgs, err := p.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (we will do it manually)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		p.logger.Error("Failed to register a consumer", "queue", q.Name, "error", err)
		return fmt.Errorf("rabbitmq: failed to register consumer: %w", err)
	}

	p.logger.Info("Subscribed to RabbitMQ queue, waiting for messages...", "queue", q.Name)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Context canceled, stopping RabbitMQ subscription loop.", "queue", q.Name)
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				p.logger.Info("RabbitMQ channel closed, stopping subscription loop.", "queue", q.Name)
				return nil
			}
			p.processMessage(ctx, d, handler)
		}
	}
}

// processMessage handles a single received AMQP delivery.
func (p *rabbitmqProvider) processMessage(ctx context.Context, d amqp.Delivery, handler HandlerFunc) {
	// Extract trace ID from headers to create a new context.
	var traceID string
	if id, ok := d.Headers[traceIDHeader].(string); ok {
		traceID = id
	}

	handlerCtx := ctxutil.WithRequestID(ctx, traceID)
	log := p.logger.With("request_id", traceID)
	handlerCtx = ctxutil.WithLogger(handlerCtx, log)

	log.Debug("Processing RabbitMQ message", "exchange", d.Exchange, "routing_key", d.RoutingKey)

	// Convert to our standard Message format.
	brokerMsg := Message{
		ID:        d.MessageId,
		Body:      d.Body,
		Timestamp: d.Timestamp,
		Headers:   make(map[string]string),
	}
	for k, v := range d.Headers {
		if s, ok := v.(string); ok {
			brokerMsg.Headers[k] = s
		}
	}

	// Execute the handler.
	if err := handler(handlerCtx, brokerMsg); err != nil {
		log.Error("Handler failed to process message", "error", err)
		// NACK the message and ask the broker to requeue it (or send to DLX).
		if err := d.Nack(false, true); err != nil {
			log.Error("Failed to NACK message", "error", err)
		}
		return
	}

	// If the handler was successful, ACK the message.
	if err := d.Ack(false); err != nil {
		log.Error("Failed to ACK message", "error", err)
	}
}

// Ping checks if the RabbitMQ connection is alive.
func (p *rabbitmqProvider) Ping(ctx context.Context) error {
	// RabbitMQ doesn't have a native Ping. A good way to check is to open a new channel.
	log := ctxutil.LoggerFromContext(ctx)
	log.Debug("Pinging RabbitMQ")
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	return ch.Close()
}

// Close gracefully closes the channel and connection.
func (p *rabbitmqProvider) Close() error {
	p.logger.Info("Closing RabbitMQ connection...")
	if err := p.ch.Close(); err != nil {
		err := p.conn.Close()
		if err != nil {
			return err
		} // Attempt to close connection even if channel close fails
		return fmt.Errorf("rabbitmq: failed to close channel: %w", err)
	}
	return p.conn.Close()
}

// Channel provides access to the native AMQP channel.
func (p *rabbitmqProvider) Channel() *amqp.Channel {
	return p.ch
}
