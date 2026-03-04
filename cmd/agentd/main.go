package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sz3/scodex/internal/api"
	"github.com/sz3/scodex/internal/app"
	"github.com/sz3/scodex/internal/ipc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	daemon, err := app.BootstrapDaemon(ctx, app.Dependencies{})
	if err != nil {
		slog.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}
	defer daemon.Shutdown()

	authToken := os.Getenv("AGENTD_AUTH_TOKEN")
	if authToken == "" {
		authToken = api.DefaultAuthToken
		daemon.Logger.Warn("using default local auth token; set AGENTD_AUTH_TOKEN for stronger local security")
	}

	handler, err := api.NewRuntimeHandler(api.Dependencies{
		Logger:    daemon.Logger,
		Sessions:  daemon.SessionStore,
		Events:    daemon.EventBus,
		AuthToken: authToken,
	})
	if err != nil {
		daemon.Logger.Error("api handler init failed", "error", err)
		os.Exit(1)
	}

	addr := os.Getenv("AGENTD_HTTP_ADDR")
	if addr == "" {
		addr = ipc.DefaultAddress
	}

	allowNonLoopback := strings.EqualFold(os.Getenv("AGENTD_ALLOW_NON_LOOPBACK"), "true")
	if err := ipc.ValidateLocalAddress(addr, allowNonLoopback); err != nil {
		daemon.Logger.Error("invalid daemon bind address", "address", addr, "error", err)
		os.Exit(1)
	}

	err = ipc.RunLocalHTTPServer(daemon.Context(), ipc.LocalHTTPConfig{
		Addr:    addr,
		Handler: handler,
		Logger:  daemon.Logger,
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		daemon.Logger.Error("daemon exited with error", "error", err)
		os.Exit(1)
	}
}
