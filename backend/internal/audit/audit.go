package audit

import (
	"context"
	"time"
)

type Event struct {
	Action    string
	ActorID   string
	Resource  string
	Timestamp time.Time
	Metadata  map[string]string
}

type Logger interface {
	Log(ctx context.Context, event Event) error
}
