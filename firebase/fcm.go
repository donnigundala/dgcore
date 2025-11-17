package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"

	"github.com/donnigundala/dgcore/ctxutil"
)

// IFCMClient defines the interface for the FCM client.
type IFCMClient interface {
	Send(ctx context.Context, token string, msg *FCMMessage) error
	SendToTopic(ctx context.Context, topic string, msg *FCMMessage) error
	SendToMultipleTokens(ctx context.Context, tokens []string, msg *FCMMessage) ([]string, error)
	SubscribeToTopic(ctx context.Context, tokens []string, topic string) error
	UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error
}

// FCMClient is a client for the Firebase Cloud Messaging service.
type FCMClient struct {
	client *messaging.Client
	logger *slog.Logger // Base logger for the client
}

// NewFCMClient creates a new FCM client from a specific Firebase app instance.
// This allows you to have FCM clients for multiple Firebase projects.
func NewFCMClient(app *fb.App, cfg *Config, logger *slog.Logger) (IFCMClient, error) {
	// Use a background context for initialization as it's a one-time setup.
	client, err := app.Messaging(context.Background())
	if err != nil {
		logger.Error("Unable to initialize FCM Client", "error", err)
		return nil, err
	}

	// Helper to extract project ID from credentials JSON for logging purposes.
	var projectID string
	if cfg.CredentialsJSON != "" {
		var creds struct {
			ProjectID string `json:"project_id"`
		}
		if err := json.Unmarshal([]byte(cfg.CredentialsJSON), &creds); err == nil {
			projectID = creds.ProjectID
		}
	}

	fcmLogger := logger.With("component", "fcm")
	if projectID != "" {
		fcmLogger = fcmLogger.With("project_id", projectID)
	}
	fcmLogger.Info("FCM client initialized")

	return &FCMClient{client: client, logger: fcmLogger}, nil
}

// FCMMessage represents the data and notification payload for an FCM message.
type FCMMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

// Send sends a push notification to a specific device token.
func (f *FCMClient) Send(ctx context.Context, token string, msg *FCMMessage) error {
	log := ctxutil.LoggerFromContext(ctx)
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Token: token,
		Data:  msg.Data,
	}

	resp, err := f.client.Send(ctx, message)
	if err != nil {
		if isInvalidTokenError(err) {
			log.Warn("Invalid or unregistered FCM token", "token", token, "error", err)
			return ErrInvalidFCMToken
		}
		log.Error("Failed to send FCM message", "token", token, "error", err)
		return err
	}

	log.Info("FCM message sent successfully", "response", resp)
	return nil
}

// SendToTopic sends a push notification to all devices subscribed to a specific topic.
func (f *FCMClient) SendToTopic(ctx context.Context, topic string, msg *FCMMessage) error {
	log := ctxutil.LoggerFromContext(ctx)
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
		log.Error("Failed to send message to FCM topic", "topic", topic, "error", err)
		return err
	}

	log.Info("Message sent to FCM topic", "topic", topic, "response", resp)
	return nil
}

// SendToMultipleTokens sends a push notification to multiple device tokens.
// It returns a slice of tokens that were found to be invalid.
func (f *FCMClient) SendToMultipleTokens(ctx context.Context, tokens []string, msg *FCMMessage) ([]string, error) {
	log := ctxutil.LoggerFromContext(ctx)
	if len(tokens) == 0 {
		return nil, errors.New("no tokens provided for FCM multicast message")
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
		log.Error("Failed to send FCM multicast message", "error", err)
		return nil, err
	}

	var invalidTokens []string
	for i, r := range resp.Responses {
		if !r.Success {
			if isInvalidTokenError(r.Error) {
				invalidTokens = append(invalidTokens, tokens[i])
				log.Warn("Invalid token detected in FCM batch", "token", tokens[i], "error", r.Error)
			} else {
				log.Error("Error sending to token in FCM batch", "token", tokens[i], "error", r.Error)
			}
		}
	}

	log.Info("FCM multicast message sent", "success_count", resp.SuccessCount, "failure_count", resp.FailureCount)
	return invalidTokens, nil
}

// SubscribeToTopic subscribes a list of device tokens to a specific topic.
func (f *FCMClient) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	log := ctxutil.LoggerFromContext(ctx)
	resp, err := f.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		log.Error("Failed to subscribe to FCM topic", "topic", topic, "error", err)
		return err
	}
	log.Info("Subscribed tokens to FCM topic", "topic", topic, "success_count", resp.SuccessCount, "failure_count", len(resp.Errors))
	return nil
}

// UnsubscribeFromTopic unsubscribes a list of device tokens from a specific topic.
func (f *FCMClient) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	log := ctxutil.LoggerFromContext(ctx)
	resp, err := f.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		log.Error("Failed to unsubscribe from FCM topic", "topic", topic, "error", err)
		return err
	}
	log.Info("Unsubscribed tokens from FCM topic", "topic", topic, "success_count", resp.SuccessCount, "failure_count", len(resp.Errors))
	return nil
}

// ErrInvalidFCMToken indicates that the provided FCM token is invalid or unregistered.
var ErrInvalidFCMToken = errors.New("invalid or unregistered FCM token")

// isInvalidTokenError checks if the error is related to an invalid or unregistered token.
func isInvalidTokenError(err error) bool {
	return messaging.IsInvalidArgument(err) || messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsUnregistered(err)
}
