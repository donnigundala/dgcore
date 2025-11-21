package broker

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nats-io/nats.go"
	amqp "github.com/rabbitmq/amqp091-go"
)

// HandlerFunc is the function signature for message handlers.
// If the function returns an error, the driver may decide to NACK (Negative Acknowledge)
// or retry processing the message, depending on the broker's capabilities.
type HandlerFunc func(ctx context.Context, msg Message) error

// Provider defines the common interface for all message brokers.
// It provides a generic abstraction for the most common messaging operations.
type Provider interface {
	// Publish sends a message to a specific topic.
	// The "topic" is a generic concept that can be a Kafka topic,
	// a RabbitMQ routing key, or a NATS subject.
	Publish(ctx context.Context, topic string, msg Message) error

	// Subscribe registers a handler for a specific topic.
	// This is typically a blocking operation and should be run in a separate goroutine.
	// The implementation should handle reconnects and retries internally.
	Subscribe(ctx context.Context, topic string, handler HandlerFunc) error

	// Ping checks if the connection to the broker is alive.
	Ping(ctx context.Context) error

	// Close gracefully closes the connection to the broker.
	Close() error
}

// KafkaProvider extends the base Provider with access to the native Kafka clients.
type KafkaProvider interface {
	Provider
	// Producer returns the underlying native Kafka producer instance for full control.
	Producer() *kafka.Producer
	// Consumer returns the underlying native Kafka consumer instance.
	Consumer() *kafka.Consumer
}

// RabbitMQProvider extends the base Provider with access to the native RabbitMQ channel.
type RabbitMQProvider interface {
	Provider
	// Channel returns the underlying AMQP channel for advanced operations,
	// such as manually declaring exchanges or queues.
	Channel() *amqp.Channel
}

// NATSProvider extends the base Provider with access to the native NATS connection.
type NATSProvider interface {
	Provider
	// Conn returns the underlying NATS connection for advanced features like JetStream.
	Conn() *nats.Conn
}

// Message is the standardized representation of a message within the framework.
type Message struct {
	// ID is a unique identifier for this message (e.g., UUID).
	ID string

	// Headers are used to carry metadata, such as trace IDs for distributed tracing.
	Headers map[string]string

	// Body is the actual payload of the message.
	Body []byte

	// Timestamp is the time the message was created.
	Timestamp time.Time
}
