package broker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/nats-io/nats.go"
)

// natsProvider implements the Provider and NATSProvider interfaces.
type natsProvider struct {
	conn   *nats.Conn
	config *Config
	logger *slog.Logger
}

// newNATSProvider creates and initializes a new NATS provider.
func newNATSProvider(cfg *Config, logger *slog.Logger) (Provider, error) {
	conn, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		logger.Error("Failed to connect to NATS", "url", cfg.NATS.URL, "error", err)
		return nil, fmt.Errorf("nats: failed to connect: %w", err)
	}

	logger.Info("NATS connection established successfully")

	return &natsProvider{
		conn:   conn,
		config: cfg,
		logger: logger,
	}, nil
}

// Publish sends a message to a NATS subject.
// The topic parameter is used as the NATS subject.
func (p *natsProvider) Publish(ctx context.Context, topic string, msg Message) error {
	log := ctxutil.LoggerFromContext(ctx)

	// Create a NATS message and set its headers for tracing.
	natsMsg := nats.NewMsg(topic)
	if traceID := ctxutil.RequestIDFromContext(ctx); traceID != "" {
		natsMsg.Header.Set(traceIDHeader, traceID)
	}
	for k, v := range msg.Headers {
		natsMsg.Header.Set(k, v)
	}
	natsMsg.Data = msg.Body

	// Publish the message.
	if err := p.conn.PublishMsg(natsMsg); err != nil {
		log.Error("Failed to publish message to NATS", "subject", topic, "error", err)
		return fmt.Errorf("nats: failed to publish: %w", err)
	}

	// For standard NATS, publish is fire-and-forget. Flushing ensures it's sent.
	if err := p.conn.Flush(); err != nil {
		log.Error("Failed to flush NATS connection after publish", "subject", topic, "error", err)
		return fmt.Errorf("nats: failed to flush: %w", err)
	}

	log.Debug("Message published successfully to NATS", "subject", topic)
	return nil
}

// Subscribe listens for messages on a NATS subject and calls the handler.
// This is a blocking operation.
func (p *natsProvider) Subscribe(ctx context.Context, topic string, handler HandlerFunc) error {
	// The NATS library handles the message loop in a separate goroutine,
	// so we create a subscription and then block until our context is canceled.
	sub, err := p.conn.Subscribe(topic, func(natsMsg *nats.Msg) {
		p.processMessage(ctx, natsMsg, handler)
	})
	if err != nil {
		p.logger.Error("Failed to subscribe to NATS subject", "subject", topic, "error", err)
		return fmt.Errorf("nats: failed to subscribe: %w", err)
	}

	p.logger.Info("Subscribed to NATS subject, waiting for messages...", "subject", topic)

	// Block until the context is done.
	<-ctx.Done()

	// Unsubscribe and drain remaining messages to ensure graceful shutdown.
	p.logger.Info("Context canceled, unsubscribing from NATS subject.", "subject", topic)
	if err := sub.Unsubscribe(); err != nil {
		p.logger.Error("Failed to unsubscribe from NATS", "subject", topic, "error", err)
	}
	if err := sub.Drain(); err != nil {
		p.logger.Error("Failed to drain NATS subscription", "subject", topic, "error", err)
	}

	return ctx.Err()
}

// processMessage handles a single received NATS message.
func (p *natsProvider) processMessage(ctx context.Context, natsMsg *nats.Msg, handler HandlerFunc) {
	// Extract trace ID from headers to create a new context.
	traceID := natsMsg.Header.Get(traceIDHeader)

	handlerCtx := ctxutil.WithRequestID(ctx, traceID)
	log := p.logger.With("request_id", traceID)
	handlerCtx = ctxutil.WithLogger(handlerCtx, log)

	log.Debug("Processing NATS message", "subject", natsMsg.Subject)

	// Convert to our standard Message format.
	brokerMsg := Message{
		ID:        natsMsg.Header.Get("Nats-Msg-Id"), // NATS generates a unique ID
		Body:      natsMsg.Data,
		Timestamp: time.Now(), // NATS msg doesn't have a timestamp field, so we use receive time
		Headers:   make(map[string]string),
	}
	for k, v := range natsMsg.Header {
		if len(v) > 0 {
			brokerMsg.Headers[k] = v[0]
		}
	}

	// Execute the handler.
	if err := handler(handlerCtx, brokerMsg); err != nil {
		log.Error("Handler failed to process message", "error", err)
		// Core NATS does not have a NACK mechanism. The message is simply dropped.
		// For guaranteed delivery, NATS JetStream would be required.
	}
}

// Ping checks if the NATS connection is alive.
func (p *natsProvider) Ping(ctx context.Context) error {
	log := ctxutil.LoggerFromContext(ctx)
	log.Debug("Pinging NATS server")

	if p.conn.Status() != nats.CONNECTED {
		return fmt.Errorf("nats: not connected, status is %s", p.conn.Status())
	}
	return nil
}

// Close gracefully closes the NATS connection.
func (p *natsProvider) Close() error {
	p.logger.Info("Closing NATS connection...")
	p.conn.Close()
	return nil
}

// Conn provides access to the native NATS connection.
func (p *natsProvider) Conn() *nats.Conn {
	return p.conn
}
