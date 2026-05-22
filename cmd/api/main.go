package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/app"
	"github.com/en7ka/hitalent_testovoe/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init application: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
