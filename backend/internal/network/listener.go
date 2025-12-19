package network

import (
	"context"
)

// Listener defines the behavior for any network protocol listener
type Listener interface {
	Start(ctx context.Context) error
	Stop() error
	Protocol() string
	Port() int
}

// Handler defines how to process incoming raw data
type Handler interface {
	Handle(data []byte, metadata map[string]string) error
}
