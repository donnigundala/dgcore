package firebase

import (
	"context"
	"errors"
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
	fb := &Firebase{}

	for _, opt := range opts {
		opt(fb)
	}

	if fb.logger == nil {
		fb.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	fb.logger = fb.logger.With("component", "firebase")

	var credOption option.ClientOption

	if cfg.CredentialJSON != "" {
		fb.logger.Info("Initializing Firebase with credentials from firebase.credentials_json environment variable")
		credOption = option.WithCredentialsJSON([]byte(cfg.CredentialJSON))
	} else if cfg.CredentialsFile != "" {
		fb.logger.Info("Initializing Firebase with credentials file", "path", cfg.CredentialsFile)
		credOption = option.WithCredentialsFile(cfg.CredentialsFile)
	} else {
		err := errors.New("firebase credentials not found. Set firebase.credentials_json or firebase.credentials_file in config")
		fb.logger.Error(err.Error())
		return nil, err
	}

	app, err := firebase.NewApp(ctx, nil, credOption)
	if err != nil {
		fb.logger.Error("Failed to initialize Firebase app", "error", err)
		return nil, err
	}

	fb.App = app
	fb.logger.Info("Initialized successfully")

	return fb, nil
}
