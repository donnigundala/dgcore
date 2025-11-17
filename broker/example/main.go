package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/donnigundala/dgcore/broker"
	"github.com/donnigundala/dgcore/config"
	"github.com/donnigundala/dgcore/ctxutil"
)

func main() {
	// 1. Bootstrap logger and load configuration
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if err := config.LoadWithPaths("./config/broker.yaml"); err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Inject broker configurations
	var brokerConfigs map[string]*broker.Config
	if err := config.Inject("brokers", &brokerConfigs); err != nil {
		logger.Error("failed to inject broker config", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the Broker Manager
	brokerManager, err := broker.NewManager(brokerConfigs, broker.WithLogger(logger))
	if err != nil {
		logger.Error("failed to initialize broker manager", "error", err)
		os.Exit(1)
	}
	defer brokerManager.Close()

	// 4. Get broker providers
	kafkaBroker, _ := brokerManager.Broker("user_events")
	rabbitBroker, _ := brokerManager.Broker("notifications")
	natsBroker, _ := brokerManager.Broker("system_logs")

	// 5. Setup context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	// 6. Start subscribers in separate goroutines
	wg.Add(3)
	go runSubscriber(ctx, &wg, kafkaBroker, "user.created")
	go runSubscriber(ctx, &wg, rabbitBroker, "email.sent")
	go runSubscriber(ctx, &wg, natsBroker, "log.info")

	// 7. Start a publisher to send messages periodically
	wg.Add(1)
	go runPublisher(ctx, &wg, kafkaBroker, rabbitBroker, natsBroker)

	// Wait for all goroutines to finish
	logger.Info("Broker example running. Press Ctrl+C to exit.")
	wg.Wait()
	logger.Info("Broker example finished.")
}

// runSubscriber starts a subscription for a given broker and topic.
func runSubscriber(ctx context.Context, wg *sync.WaitGroup, b broker.Provider, topic string) {
	defer wg.Done()
	logger := slog.Default().With("subscriber_topic", topic)
	logger.Info("Starting subscriber...")

	handler := func(handlerCtx context.Context, msg broker.Message) error {
		log := ctxutil.LoggerFromContext(handlerCtx)
		log.Info("Received message", "id", msg.ID, "body", string(msg.Body))
		return nil
	}

	// Subscribe is a blocking call, it will exit when the context is canceled.
	if err := b.Subscribe(ctx, topic, handler); err != nil {
		logger.Error("Subscriber stopped with error", "error", err)
	} else {
		logger.Info("Subscriber stopped gracefully.")
	}
}

// runPublisher periodically sends messages to all brokers.
func runPublisher(ctx context.Context, wg *sync.WaitGroup, brokers ...broker.Provider) {
	defer wg.Done()
	logger := slog.Default().With("component", "publisher")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	msgCounter := 0
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping publisher...")
			return
		case <-ticker.C:
			msgCounter++
			// Create a new context for each "request" or "transaction"
			pubCtx := ctxutil.WithRequestID(context.Background(), fmt.Sprintf("req-%d", msgCounter))
			pubCtx = ctxutil.WithLogger(pubCtx, logger.With("request_id", ctxutil.RequestIDFromContext(pubCtx)))

			log := ctxutil.LoggerFromContext(pubCtx)

			// Publish to Kafka
			log.Info("Publishing to Kafka...")
			kafkaMsg := broker.Message{Body: []byte(fmt.Sprintf("Kafka message #%d", msgCounter))}
			if err := brokers[0].Publish(pubCtx, "user.created", kafkaMsg); err != nil {
				log.Error("Failed to publish to Kafka", "error", err)
			}

			// Publish to RabbitMQ
			log.Info("Publishing to RabbitMQ...")
			rabbitMsg := broker.Message{Body: []byte(fmt.Sprintf("RabbitMQ message #%d", msgCounter))}
			if err := brokers[1].Publish(pubCtx, "email.sent", rabbitMsg); err != nil {
				log.Error("Failed to publish to RabbitMQ", "error", err)
			}

			// Publish to NATS
			log.Info("Publishing to NATS...")
			natsMsg := broker.Message{Body: []byte(fmt.Sprintf("NATS message #%d", msgCounter))}
			if err := brokers[2].Publish(pubCtx, "log.info", natsMsg); err != nil {
				log.Error("Failed to publish to NATS", "error", err)
			}
		}
	}
}
