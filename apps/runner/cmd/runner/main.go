package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/samuelsih/golib/slogx"
	"github.com/samuelsih/nibiru/runner"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("services", "nibiru-runner"),
	)
	slog.SetDefault(logger)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("Error when running application", slogx.ErrorAttr(err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := env.ParseAs[runner.Config]()
	if err != nil {
		return fmt.Errorf("cannot parse config: %w", err)
	}

	rootCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	js, consumer, err := cfg.ConnectPubSub(rootCtx)
	if err != nil {
		return fmt.Errorf("cannot connect pubsub: %w", err)
	}
	defer js.Conn().Close()

	go runner.Consume(rootCtx, consumer)

	<-rootCtx.Done()

	slog.Info("Received termination signal, shutting down")

	return nil
}
