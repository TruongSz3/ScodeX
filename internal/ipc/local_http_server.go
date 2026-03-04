package ipc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const DefaultAddress = "127.0.0.1:7777"

var ErrNonLoopbackBind = errors.New("ipc: non-loopback bind requires AGENTD_ALLOW_NON_LOOPBACK=true")

type LocalHTTPConfig struct {
	Addr    string
	Handler http.Handler
	Logger  *slog.Logger
}

func ValidateLocalAddress(addr string, allowNonLoopback bool) error {
	if allowNonLoopback {
		return nil
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("ipc: invalid listen address %q: %w", addr, err)
	}

	if host == "localhost" {
		return nil
	}

	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}

	return fmt.Errorf("%w (%s)", ErrNonLoopbackBind, addr)
}

func RunLocalHTTPServer(ctx context.Context, cfg LocalHTTPConfig) error {
	if ctx == nil {
		return errors.New("ipc: context is required")
	}
	if cfg.Handler == nil {
		return errors.New("ipc: handler is required")
	}

	addr := cfg.Addr
	if addr == "" {
		addr = DefaultAddress
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger := cfg.Logger
	if logger != nil {
		logger.Info("local HTTP fallback listening", "addr", ln.Addr().String())
	}

	server := &http.Server{
		Handler:           cfg.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil && logger != nil {
			logger.Error("local HTTP fallback shutdown failed", "error", shutdownErr)
		}
		serveResult := <-serveErr
		if serveResult != nil && !errors.Is(serveResult, http.ErrServerClosed) {
			return serveResult
		}
		return nil
	case err := <-serveErr:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
