package shutdown

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Task = func(ctx context.Context) error

func Shutdown(ctx context.Context, timeout time.Duration, operations map[string]Task) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

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
			}()
		}
		wg.Wait()
		close(wait)
	}()
	return wait
}
