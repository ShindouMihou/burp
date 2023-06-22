package main

import (
	"burp/internal/app"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ShutdownTask = func(ctx context.Context) error

func main() {
	app.Init()
	// TODO: Add CLI application logic here.
	<-Shutdown(context.Background(), 5*time.Second, map[string]ShutdownTask{
		"cleanup_burp": func(ctx context.Context) error {
			return Cleanup()
		},
	})
	return
}

func Shutdown(ctx context.Context, timeout time.Duration, operations map[string]ShutdownTask) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Info().Msg("Preparing to shutdown burp-agent...")
		out := time.AfterFunc(timeout, func() {
			log.Error().Msg("Graceful shutdown couldn't be completed even after timeout, forcing...")
		})
		defer out.Stop()
		var wg sync.WaitGroup
		for key, operation := range operations {
			wg.Add(1)

			key := key
			operation := operation
			go func() {
				defer wg.Done()
				if err := operation(ctx); err != nil {
					log.Err(err).Str("task", key).Msg("Failed to complete shutdown task")
					return
				}
				log.Info().Str("task", key).Msg("Shutdown task was completed successfully.")
			}()
		}
		wg.Wait()
		close(wait)
	}()
	return wait
}

func Cleanup() error {
	if err := os.RemoveAll(".burp/"); err != nil {
		return errors.Join(errors.New("failed to cleanup .burp/ folder"), err)
	}
	return nil
}
