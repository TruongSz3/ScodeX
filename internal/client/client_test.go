package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	cli := New(WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	resp, err := cli.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("check health failed: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestCheckHealthNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cli := New(WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := cli.CheckHealth(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPingHandshakeSequence(t *testing.T) {
	var (
		orderMu sync.Mutex
		order   []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderMu.Lock()
		order = append(order, r.URL.Path)
		orderMu.Unlock()

		switch r.URL.Path {
		case "/runtime/initialize":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method for initialize: %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(InitializeResponse{SessionID: "s-123"})
		case "/runtime/initialized":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method for initialized: %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(InitializedResponse{Status: "ok"})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cli := New(WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err := cli.PingHandshake(context.Background()); err != nil {
		t.Fatalf("ping handshake failed: %v", err)
	}

	orderMu.Lock()
	defer orderMu.Unlock()
	if len(order) != 2 {
		t.Fatalf("unexpected call count: %d", len(order))
	}
	if order[0] != "/runtime/initialize" || order[1] != "/runtime/initialized" {
		t.Fatalf("unexpected order: %#v", order)
	}
}
