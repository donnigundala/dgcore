package firebase

import (
	"context"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// Firebase wraps the Firebase app object.
type Firebase struct {
	App    *firebase.App
	logger *slog.Logger
}

// Option defines a functional option for configuring the Firebase instance.
type Option func(*Firebase)

// WithLogger sets a custom logger for the Firebase instance.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Firebase) {
		s.logger = logger
	}
}

// NewFirebase initializes and returns a new Firebase instance.
func NewFirebase(ctx context.Context, cfg *Config, opts ...Option) (*Firebase, error) {
	opt := option.WithCredentialsFile(cfg.CredentialsFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		// Temporary logger for initialization errors
		tempLogger := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "firebase")
		tempLogger.Error("Failed to initialize Firebase app", "error", err)
		return nil, err
	}

	fb := &Firebase{App: app}

	for _, opt := range opts {
		opt(fb)
	}

	if fb.logger == nil {
		fb.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	fb.logger = fb.logger.With("component", "firebase")
	fb.logger.Info("Initialized successfully")

	return fb, nil
}
