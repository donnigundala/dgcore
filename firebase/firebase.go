package firebase

import (
	"context"
	"errors"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// firebaseApp wraps the Firebase app object.
type firebaseApp struct {
	App    *firebase.App
	logger *slog.Logger
}

// appOption defines a functional option for configuring the Firebase instance.
type appOption func(*firebaseApp)

// withAppLogger sets a custom logger for the Firebase instance.
func withAppLogger(logger *slog.Logger) appOption {
	return func(s *firebaseApp) {
		s.logger = logger
	}
}

// newFirebaseApp initializes and returns a new Firebase instance.
func newFirebaseApp(ctx context.Context, cfg *Config, opts ...appOption) (*firebaseApp, error) {
	fb := &firebaseApp{}

	for _, opt := range opts {
		opt(fb)
	}

	if fb.logger == nil {
		fb.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	fb.logger = fb.logger.With("component", "firebase")

	var credOption option.ClientOption

	if cfg.CredentialsJSON != "" {
		fb.logger.Info("Initializing Firebase with credentials from firebase.credentials_json environment variable")
		credOption = option.WithCredentialsJSON([]byte(cfg.CredentialsJSON))
	} else if cfg.CredentialsFile != "" {
		fb.logger.Info("Initializing Firebase with credentials file", "path", cfg.CredentialsFile)
		credOption = option.WithCredentialsFile(cfg.CredentialsFile)
	} else {
		err := errors.New("firebase credentials not found. Set firebase.credentials_json or firebase.credentials_file in your config")
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
