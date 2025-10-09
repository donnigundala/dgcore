package firebase

import (
	"context"
	"errors"
	"log"
	"strings"

	"firebase.google.com/go/v4/messaging"
)

type FCMClient struct {
	client *messaging.Client
}

func NewFCM(firebase *Firebase) *FCMClient {
	client, err := firebase.App.Messaging(context.Background())
	if err != nil {
		log.Fatalf("[FCM] Failed to initialize messaging client: %v", err)
	}
	log.Println("[FCM] Messaging client initialized")
	return &FCMClient{client: client}
}

// -----------------------------------------------------
// Message struct for single-target messages
// -----------------------------------------------------

// FCMMessage represents the structure of a message to be sent via FCM.
type FCMMessage struct {
	Title string
	Body  string
	Token string
	Data  map[string]string
}

// -----------------------------------------------------
// Send single message
// -----------------------------------------------------

// Send sends a push notification to a specific device token.
func (f *FCMClient) Send(ctx context.Context, msg FCMMessage) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Token: msg.Token,
		Data:  msg.Data,
	}

	resp, err := f.client.Send(ctx, message)
	if err != nil {
		// Handle invalid token errors
		if isInvalidTokenError(err) {
			log.Printf("[FCM] Invalid or unregistered token: %s", msg.Token)
			return ErrInvalidToken
		}
		return err
	}

	log.Printf("[FCM] Message sent successfully: %s", resp)
	return nil
}

// -----------------------------------------------------
// Broadcast message to a topic (e.g., "news_updates")
// -----------------------------------------------------

// SendToTopic sends a push notification to all devices subscribed to a specific topic.
func (f *FCMClient) SendToTopic(ctx context.Context, topic string, msg FCMMessage) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Topic: topic,
		Data:  msg.Data,
	}

	resp, err := f.client.Send(ctx, message)
	if err != nil {
		return err
	}

	log.Printf("[FCM] Message sent to topic '%s': %s", topic, resp)
	return nil
}

// -----------------------------------------------------
// Broadcast to multiple tokens (up to 500 tokens)
// -----------------------------------------------------

// SendToMultipleTokens sends a push notification to multiple device tokens.
func (f *FCMClient) SendToMultipleTokens(ctx context.Context, msg FCMMessage, tokens []string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens provided")
	}

	batch := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Tokens: tokens,
		Data:   msg.Data,
	}

	resp, err := f.client.SendEachForMulticast(ctx, batch)
	if err != nil {
		return nil, err
	}

	var invalidTokens []string
	for i, r := range resp.Responses {
		if !r.Success {
			if isInvalidTokenError(r.Error) {
				invalidTokens = append(invalidTokens, tokens[i])
				log.Printf("[FCM] Invalid token detected: %s", tokens[i])
			} else {
				log.Printf("[FCM] Error sending to %s: %v", tokens[i], r.Error)
			}
		}
	}

	log.Printf("[FCM] Successfully sent %d messages out of %d", resp.SuccessCount, len(tokens))
	return invalidTokens, nil
}

// -----------------------------------------------------
// Subscribe / Unsubscribe tokens to topic
// -----------------------------------------------------

// SubscribeToTopic subscribes a list of device tokens to a specific topic.
func (f *FCMClient) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	resp, err := f.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return err
	}
	log.Printf("[FCM] Subscribed %d tokens to topic '%s'", resp.SuccessCount, topic)
	return nil
}

func (f *FCMClient) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	resp, err := f.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return err
	}
	log.Printf("[FCM] Unsubscribed %d tokens from topic '%s'", resp.SuccessCount, topic)
	return nil
}

// -----------------------------------------------------
// Error Handling Helpers
// -----------------------------------------------------

// ErrInvalidToken indicates that the provided FCM token is invalid or unregistered.
var ErrInvalidToken = errors.New("invalid or unregistered FCM token")

// isInvalidTokenError checks if the error is related to an invalid or unregistered token.
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "registration token is not a valid FCM registration token") ||
		strings.Contains(msg, "Requested entity was not found") ||
		strings.Contains(msg, "Unregistered")
}
