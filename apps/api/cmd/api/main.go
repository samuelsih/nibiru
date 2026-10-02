package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/slogx"
	"github.com/samuelsih/nibiru/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("services", "nibiru-api"),
	)
	slog.SetDefault(logger)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("Error when running application", slogx.ErrorAttr(err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	conf, err := env.ParseAs[api.Config]()
	if err != nil {
		return fmt.Errorf("cannot parse env: %w", err)
	}

	rootCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := conf.ConnectDB(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect db: %w", err)
	}

	defer db.Close()

	if err = conf.MigrationUp(ctx, db); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	httpHandler := conf.Webserver()
	httpHandler.Use(
		httpx.MiddlewareRequestID(),
		httpx.MiddlewareTimeout(conf.ServerRequestTimeout),
	)

	ongoingCtx, stopOngoing := context.WithCancel(ctx)
	defer stopOngoing()

	httpServer := http.Server{
		Addr:              conf.HostPort(),
		Handler:           httpHandler,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		ReadHeaderTimeout: conf.ServerReadHeaderTimeout,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	errChan := make(chan error, 1)

	go func() {
		slog.Info("Starting http server", slog.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("http server error: %w", err)
	case <-rootCtx.Done():
		stop()
	}

	slog.Info("Received termination signal, shutting down")

	shutdownCtx, cancel := context.WithTimeout(ctx, conf.ServerShutdownTimeout)
	err = httpServer.Shutdown(shutdownCtx)
	cancel()
	stopOngoing()

	if err != nil {
		slog.Warn("Shutdown timed out, forcing close", slogx.ErrorAttr(err))
		time.Sleep(conf.ServerShutdownHardTimeout)

		if closeErr := httpServer.Close(); closeErr != nil {
			slog.Warn("Failed to force-close the http server", slogx.ErrorAttr(closeErr))
		}
	}

	slog.Info("Server shut down")

	return nil
}
