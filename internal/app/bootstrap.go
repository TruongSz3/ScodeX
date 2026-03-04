package app

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/sz3/scodex/internal/events"
	"github.com/sz3/scodex/internal/session"
)

var ErrContextRequired = errors.New("app: context is required")

type Dependencies struct {
	Logger       *slog.Logger
	EventBus     *events.Bus
	SessionStore *session.Store
}

type Daemon struct {
	ctx          context.Context
	cancel       context.CancelFunc
	Logger       *slog.Logger
	EventBus     *events.Bus
	SessionStore *session.Store
}

func BootstrapDaemon(parent context.Context, deps Dependencies) (*Daemon, error) {
	if parent == nil {
		return nil, ErrContextRequired
	}

	ctx, cancel := context.WithCancel(parent)

	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	bus := deps.EventBus
	if bus == nil {
		bus = events.NewBus(events.DefaultSubscriberBuffer)
	}

	store := deps.SessionStore
	if store == nil {
		store = session.NewStore()
	}

	return &Daemon{
		ctx:          ctx,
		cancel:       cancel,
		Logger:       logger,
		EventBus:     bus,
		SessionStore: store,
	}, nil
}

func (d *Daemon) Context() context.Context {
	return d.ctx
}

func (d *Daemon) Shutdown() {
	if d.cancel != nil {
		d.cancel()
	}
}
