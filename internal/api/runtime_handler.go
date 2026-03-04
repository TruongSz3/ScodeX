package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/sz3/scodex/internal/events"
	"github.com/sz3/scodex/internal/session"
)

const (
	sessionCreatedTopic = "session.created"
	maxJSONBodySize     = 1 << 20
)

var ErrHandshakeIncomplete = errors.New("runtime handshake required: call POST /runtime/initialize then POST /runtime/initialized")

type SessionStore interface {
	Create(ctx context.Context, id string) (session.Session, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) (events.PublishResult, error)
}

type Dependencies struct {
	Logger    *slog.Logger
	Sessions  SessionStore
	Events    EventPublisher
	AuthToken string
}

type RuntimeHandler struct {
	logger    *slog.Logger
	sessions  SessionStore
	events    EventPublisher
	handshake *runtimeHandshake
	authToken string
	mux       *http.ServeMux
}

func NewRuntimeHandler(deps Dependencies) (*RuntimeHandler, error) {
	if deps.Sessions == nil {
		return nil, errors.New("api: sessions dependency is required")
	}
	if deps.Events == nil {
		return nil, errors.New("api: events dependency is required")
	}

	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	authToken := deps.AuthToken
	if authToken == "" {
		authToken = DefaultAuthToken
	}

	h := &RuntimeHandler{
		logger:    logger,
		sessions:  deps.Sessions,
		events:    deps.Events,
		handshake: &runtimeHandshake{},
		authToken: authToken,
		mux:       http.NewServeMux(),
	}

	h.registerRoutes()
	return h, nil
}

func (h *RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *RuntimeHandler) registerRoutes() {
	h.mux.HandleFunc("GET /health", h.handleHealth)
	h.mux.HandleFunc("GET /ready", h.handleReady)
	h.mux.HandleFunc("POST /runtime/initialize", h.handleInitialize)
	h.mux.HandleFunc("POST /runtime/initialized", h.handleInitialized)
	h.mux.HandleFunc("POST /session/create", h.handleSessionCreate)
}

func (h *RuntimeHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *RuntimeHandler) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !h.handshake.Ready() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"error":  ErrHandshakeIncomplete.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func (h *RuntimeHandler) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}

	var req initializeRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	handshakeSession, err := h.handshake.MarkInitialize(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProtocolVersionRequired), errors.Is(err, ErrProtocolVersionUnsupported):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to initialize runtime handshake"})
		}
		return
	}

	h.logger.Info("runtime initialize acknowledged")
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"phase":           "initialize",
		"sessionId":       handshakeSession,
		"protocolVersion": SupportedProtocolVersion,
	})
}

func (h *RuntimeHandler) handleInitialized(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}

	var req initializedRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	if err := h.handshake.MarkInitialized(req); err != nil {
		switch {
		case errors.Is(err, ErrHandshakeSessionRequired):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		case errors.Is(err, ErrInitializeRequired), errors.Is(err, ErrHandshakeSessionMismatch):
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to complete runtime handshake"})
		}
		return
	}

	h.logger.Info("runtime handshake complete")
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "phase": "initialized"})
}

type createSessionRequest struct {
	ID string `json:"id"`
}

func (h *RuntimeHandler) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}

	if !h.handshake.Ready() {
		writeJSON(w, http.StatusPreconditionRequired, map[string]any{"error": ErrHandshakeIncomplete.Error()})
		return
	}

	var req createSessionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	created, err := h.sessions.Create(r.Context(), req.ID)
	if err != nil {
		switch {
		case errors.Is(err, session.ErrSessionIDRequired):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		case errors.Is(err, session.ErrSessionExists):
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create session"})
		}
		return
	}

	response := map[string]any{
		"id":              created.ID,
		"created_at":      created.CreatedAt.Format(time.RFC3339),
		"event_published": true,
	}

	if _, err := h.events.Publish(r.Context(), sessionCreatedTopic, map[string]any{"id": created.ID}); err != nil {
		h.logger.Warn("session created but event publish failed", "session_id", created.ID, "error", err)
		response["event_published"] = false
		response["event_warning"] = "session created but event publish failed"
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *RuntimeHandler) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get(HeaderAuthToken) != h.authToken {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized: invalid or missing auth token"})
		return false
	}
	return true
}

func decodeJSONBody(r *http.Request, out any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodySize))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values are not allowed")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
