# `broker` Package

## Overview

The `broker` package provides a unified and high-performance interface for interacting with various message brokers like Kafka, RabbitMQ, and NATS. It is designed for robustness and ease of use, integrating seamlessly with the framework's configuration and context-aware logging systems.

## Core Concepts

### 1. The `Manager`

The `Manager` is the central component that handles the lifecycle of all named broker connections. It is initialized from your application's configuration and provides a single point of access to all broker providers.

### 2. The `Provider` Interface

All broker drivers (Kafka, RabbitMQ, NATS) implement a consistent `Provider` interface for the most common operations: `Publish` and `Subscribe`. This allows you to write application code that is agnostic to the underlying broker technology.

```go
type Provider interface {
    Publish(ctx context.Context, topic string, msg Message) error
    Subscribe(ctx context.Context, topic string, handler HandlerFunc) error
    Ping(ctx context.Context) error
    Close() error
}
```

### 3. Standardized `Message` Struct

To enable powerful features like distributed tracing, all operations use a standardized `Message` struct. The `Headers` map is critical for passing metadata, such as a `request_id`, across service boundaries.

```go
type Message struct {
    ID        string
    Headers   map[string]string
    Body      []byte
    Timestamp time.Time
}
```

### 4. Context-Aware Operations & Distributed Tracing

Every operation is context-aware. This is the key to building observable, distributed systems.
- **`Publish(ctx, ...)`**: The context is used to derive a context-aware logger. Crucially, it also injects a `trace_id` (from the context's `request_id`) into the message headers before sending.
- **`Subscribe(ctx, ...)`**: When a message is received, the driver extracts the `trace_id` from the message headers and creates a new context for the handler. This provides seamless, end-to-end observability across different services.

### 5. "Escape Hatch" for Native Features

While the `Provider` interface covers common cases, you can access the full power of the underlying broker client by type-asserting to a more specific interface (`KafkaProvider`, `RabbitMQProvider`, `NATSProvider`).

```go
// Get the generic provider
brokerProvider, _ := manager.Broker("my_kafka")

// Type-assert to access native features
if kafkaProvider, ok := brokerProvider.(broker.KafkaProvider); ok {
    // Now you have access to the native client
    nativeProducer := kafkaProvider.Producer()
    nativeProducer.Flush(15 * 1000)
}
```

## Configuration Example

Define your broker connections in your application's configuration file.

```yaml
# config/broker.yaml
brokers:
  # A Kafka connection
  user_events:
    driver: "kafka"
    kafka:
      brokers:
        - "localhost:9092"
      group_id: "my_app_group"

  # A RabbitMQ connection
  notifications:
    driver: "rabbitmq"
    rabbitmq:
      url: "amqp://guest:guest@localhost:5672/"
      exchange: "app_exchange"

  # A NATS connection
  system_logs:
    driver: "nats"
    nats:
      url: "nats://localhost:4222"
```

## Full Usage Example

The `broker/example/main.go` file provides a complete, runnable example that demonstrates how to publish and subscribe to all three supported brokers simultaneously.

To run the example, you will need Kafka, RabbitMQ, and NATS running locally (e.g., via Docker). Then, you can run the main file:

```bash
go run core/broker/example/main.go
```

The example will:
1. Initialize the `BrokerManager` with the three configured connections.
2. Start three separate `Subscriber` goroutines, one for each broker.
3. Start a `Publisher` goroutine that periodically sends a message to all three brokers. Each message will have a unique `request_id`.
4. The subscribers will receive the messages and log them, demonstrating that the `request_id` is correctly propagated from the publisher to the subscriber, enabling distributed tracing.
5. The application will shut down gracefully when you press `Ctrl+C`.
