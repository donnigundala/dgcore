package memcache

import (
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	provider2 "github.com/donnigundala/dgcore/cache/provider"
)

type Memcache struct {
	client *memcache.Client
	config *provider2.Config
}

// New creates and initializes a Memcache client using the provided configuration.
func New(cfg *provider2.Config) (*Memcache, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mc := memcache.New(addr)
	// try a test ping
	if err := mc.Set(&memcache.Item{Key: "ping", Value: []byte("pong"), Expiration: 1}); err != nil {
		return nil, fmt.Errorf("failed to connect to memcache: %w", err)
	}
	return &Memcache{client: mc, config: cfg}, nil
}
