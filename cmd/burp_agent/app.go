package main

import (
	"burp/internal/docker"
	"burp/internal/ecosystem"
	"burp/internal/server"
	"burp/pkg/shutdown"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type ShutdownTask = func(ctx context.Context) error

func main() {
	ecosystem.Init()
	if err := docker.Init(); err != nil {
		log.Panic().Err(err).Msg("Cannot connect to Docker")
	}
	server.Init()
	<-shutdown.Shutdown(context.Background(), 5*time.Second, map[string]ShutdownTask{
		"cleanup_burp": func(ctx context.Context) error {
			return Cleanup()
		},
	})
}

func Cleanup() error {
	if err := os.RemoveAll(".burp/"); err != nil {
		return errors.Join(errors.New("failed to cleanup .burp/ folder"), err)
	}
	return nil
}
