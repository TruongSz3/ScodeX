package session

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrSessionIDRequired = errors.New("session: session id is required")
	ErrSessionExists     = errors.New("session: session already exists")
)

type Session struct {
	ID        string
	CreatedAt time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewStore() *Store {
	return &Store{sessions: make(map[string]Session)}
}

func (s *Store) Create(ctx context.Context, id string) (Session, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return Session{}, ctx.Err()
		default:
		}
	}
	if id == "" {
		return Session{}, ErrSessionIDRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; exists {
		return Session{}, ErrSessionExists
	}

	created := Session{ID: id, CreatedAt: time.Now().UTC()}
	s.sessions[id] = created

	return created, nil
}

func (s *Store) Get(ctx context.Context, id string) (Session, bool, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return Session{}, false, ctx.Err()
		default:
		}
	}
	if id == "" {
		return Session{}, false, ErrSessionIDRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	return session, ok, nil
}
