package api

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	HeaderAuthToken          = "X-Agent-Token"
	DefaultAuthToken         = "local-dev-token"
	SupportedProtocolVersion = "v1"
)

var (
	ErrInitializeRequired         = errors.New("runtime initialize required before POST /runtime/initialized")
	ErrProtocolVersionRequired    = errors.New("runtime initialize requires protocolVersion")
	ErrProtocolVersionUnsupported = errors.New("runtime initialize protocolVersion is not supported")
	ErrHandshakeSessionRequired   = errors.New("runtime initialized requires sessionId")
	ErrHandshakeSessionMismatch   = errors.New("runtime initialized sessionId does not match initialize response")
)

type initializeRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientName      string `json:"clientName"`
}

type initializedRequest struct {
	SessionID string `json:"sessionId"`
}

type runtimeHandshake struct {
	mu               sync.RWMutex
	initialized      bool
	ready            bool
	protocolVersion  string
	handshakeSession string
}

func (s *runtimeHandshake) MarkInitialize(req initializeRequest) (string, error) {
	version := strings.TrimSpace(req.ProtocolVersion)
	if version == "" {
		return "", ErrProtocolVersionRequired
	}
	if version != SupportedProtocolVersion {
		return "", ErrProtocolVersionUnsupported
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.initialized = true
	s.ready = false
	s.protocolVersion = version
	s.handshakeSession = makeHandshakeSessionID()

	return s.handshakeSession, nil
}

func (s *runtimeHandshake) MarkInitialized(req initializedRequest) error {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return ErrHandshakeSessionRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return ErrInitializeRequired
	}
	if sessionID != s.handshakeSession {
		return ErrHandshakeSessionMismatch
	}

	s.ready = true
	return nil
}

func (s *runtimeHandshake) Ready() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

func makeHandshakeSessionID() string {
	return fmt.Sprintf("hs-%d", time.Now().UTC().UnixNano())
}
