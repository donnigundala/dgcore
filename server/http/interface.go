package http

import (
	"context"
)

type IServer interface {
	Start() error
	Close() error
	Addr() string
	Shutdown(ctx context.Context) error
}
