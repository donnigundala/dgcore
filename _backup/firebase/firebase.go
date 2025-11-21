package firebase

import (
	"context"
	"errors"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// firebaseApp wraps the main Firebase app object.
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

// newFirebaseApp initializes and returns a new Firebase app instance.
func newFirebaseApp(ctx context.Context, cfg *Config, opts ...appOption) (*firebaseApp, error) {
	fb := &firebaseApp{}

	for _, opt := range opts {
		opt(fb)
	}

	// If no logger was provided by the options (e.g., from the manager),
	// create a default one.
	if fb.logger == nil {
		fb.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	fb.logger = fb.logger.With("component", "firebase")

	var credOption option.ClientOption

	// Determine the credentials source.
	if cfg.CredentialsJSON != "" {
		fb.logger.Info("Initializing Firebase with credentials from 'credentials_json' config key")
		credOption = option.WithCredentialsJSON([]byte(cfg.CredentialsJSON))
	} else if cfg.CredentialsFile != "" {
		fb.logger.Info("Initializing Firebase with credentials file", "path", cfg.CredentialsFile)
		credOption = option.WithCredentialsFile(cfg.CredentialsFile)
	} else {
		err := errors.New("firebase credentials not found. Set 'credentials_json' or 'credentials_file' in your config")
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
