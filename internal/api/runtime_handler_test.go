package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sz3/scodex/internal/events"
	"github.com/sz3/scodex/internal/session"
)

func TestRuntimeHandshakeAndSessionCreate(t *testing.T) {
	t.Parallel()

	h, err := NewRuntimeHandler(Dependencies{
		Sessions: session.NewStore(),
		Events:   events.NewBus(events.DefaultSubscriberBuffer),
	})
	if err != nil {
		t.Fatalf("NewRuntimeHandler returned error: %v", err)
	}

	t.Run("post endpoints require auth token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/runtime/initialize", bytes.NewBufferString(`{"protocolVersion":"v1"}`))

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("ready endpoint requires handshake", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected %d got %d", http.StatusServiceUnavailable, rec.Code)
		}
	})

	t.Run("initialized requires initialize first", func(t *testing.T) {
		h2, err := NewRuntimeHandler(Dependencies{
			Sessions: session.NewStore(),
			Events:   events.NewBus(events.DefaultSubscriberBuffer),
		})
		if err != nil {
			t.Fatalf("NewRuntimeHandler returned error: %v", err)
		}

		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodPost, "/runtime/initialized", bytes.NewBufferString(`{"sessionId":"s1"}`))

		h2.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected %d got %d", http.StatusConflict, rec.Code)
		}
	})

	t.Run("session create blocked until handshake completes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodPost, "/session/create", bytes.NewBufferString(`{"id":"s1"}`))

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusPreconditionRequired {
			t.Fatalf("expected %d got %d", http.StatusPreconditionRequired, rec.Code)
		}
	})

	t.Run("initialize then initialized then create session", func(t *testing.T) {
		initRec := httptest.NewRecorder()
		initReq := authedRequest(http.MethodPost, "/runtime/initialize", bytes.NewBufferString(`{"protocolVersion":"v1","clientName":"test"}`))
		h.ServeHTTP(initRec, initReq)
		if initRec.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, initRec.Code)
		}

		var initBody map[string]any
		if err := json.Unmarshal(initRec.Body.Bytes(), &initBody); err != nil {
			t.Fatalf("failed to decode initialize response: %v", err)
		}
		handshakeSession, ok := initBody["sessionId"].(string)
		if !ok || handshakeSession == "" {
			t.Fatalf("expected non-empty sessionId in initialize response, got %+v", initBody)
		}

		initializedRec := httptest.NewRecorder()
		initializedReq := authedRequest(http.MethodPost, "/runtime/initialized", bytes.NewBufferString(`{"sessionId":"`+handshakeSession+`"}`))
		h.ServeHTTP(initializedRec, initializedReq)
		if initializedRec.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, initializedRec.Code)
		}

		readyRec := httptest.NewRecorder()
		readyReq := httptest.NewRequest(http.MethodGet, "/ready", nil)
		h.ServeHTTP(readyRec, readyReq)
		if readyRec.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, readyRec.Code)
		}

		body, err := json.Marshal(map[string]string{"id": "session-1"})
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}

		createRec := httptest.NewRecorder()
		createReq := authedRequest(http.MethodPost, "/session/create", bytes.NewReader(body))
		h.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("expected %d got %d", http.StatusCreated, createRec.Code)
		}
	})
}

func authedRequest(method, path string, body io.Reader) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(HeaderAuthToken, DefaultAuthToken)
	req.Header.Set("Content-Type", "application/json")
	return req
}
