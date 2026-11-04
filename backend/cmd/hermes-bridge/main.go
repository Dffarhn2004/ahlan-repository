//go:build linux

// Command hermes-bridge is the only process that talks to the Docker CLI. It listens on a Unix domain socket, serves
// a fixed allow-list of operations for ONE container, and validates every parameter. The API and the worker call it
// instead of Docker, so they need no Docker socket, no docker group and no sudo.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ponpes-saas/internal/hermes"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log); err != nil {
		log.Error("hermes bridge stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	socket := env("HERMES_BRIDGE_SOCKET", "/run/ahlan-hermes/bridge.sock")
	container := env("HERMES_CONTAINER", "hermes-ahlan")
	mode, err := strconv.ParseUint(env("HERMES_BRIDGE_SOCKET_MODE", "0660"), 8, 32)
	if err != nil || mode&^0o770 != 0 {
		return errors.New("HERMES_BRIDGE_SOCKET_MODE must be an octal mode without world access, e.g. 0660")
	}
	uids, err := parseUIDs(os.Getenv("HERMES_BRIDGE_ALLOWED_UIDS"))
	if err != nil {
		return err
	}
	server, err := hermes.NewBridgeServer(hermes.BridgeServerConfig{
		Container:       container,
		AllowedEventURL: strings.TrimSpace(os.Getenv("HERMES_EVENTS_URL")),
		AllowedUIDs:     uids,
		Logger:          log,
	}, hermes.NewDockerExecutor(container))
	if err != nil {
		return err
	}

	// A stale socket from a previous run would make Listen fail; only a socket file is ever removed.
	if info, statErr := os.Lstat(socket); statErr == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return errors.New("HERMES_BRIDGE_SOCKET exists and is not a socket")
		}
		_ = os.Remove(socket)
	}
	// Create the socket with no access for others from the first instant, then widen to the configured group access.
	previous := syscall.Umask(0o077)
	listener, err := net.Listen("unix", socket)
	syscall.Umask(previous)
	if err != nil {
		return err
	}
	if err := os.Chmod(socket, os.FileMode(mode)); err != nil {
		return err
	}

	httpServer := &http.Server{
		Handler:           server.Handler(),
		ConnContext:       server.ConnContext,
		ReadHeaderTimeout: 10 * time.Second,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()
	notifyReady(log)
	log.Info("hermes bridge ready", "socket", socket, "container", container, "allowed_uids", len(uids))
	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func parseUIDs(raw string) ([]uint32, error) {
	var uids []uint32
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		uid, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, errors.New("HERMES_BRIDGE_ALLOWED_UIDS must be a comma-separated list of numeric UIDs")
		}
		uids = append(uids, uint32(uid))
	}
	return uids, nil
}

// notifyReady tells systemd (Type=notify) that the socket is listening, so units ordered After= us never race it.
func notifyReady(log *slog.Logger) {
	path := os.Getenv("NOTIFY_SOCKET")
	if path == "" {
		return
	}
	conn, err := net.Dial("unixgram", path)
	if err != nil {
		log.Warn("sd_notify failed", "error", err)
		return
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("READY=1"))
}
