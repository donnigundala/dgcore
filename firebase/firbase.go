package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Firebase struct {
	App *firebase.App
}

func NewFirebase(ctx context.Context, cfg *Config) *Firebase {
	opt := option.WithCredentialsFile(cfg.CredentialsFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("[Firebase] Failed to initialize: %v", err)
	}

	log.Println("[Firebase] Initialized successfully")
	return &Firebase{App: app}
}
