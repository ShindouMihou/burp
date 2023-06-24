package main

import (
	"burp/cmd/burp-agent/server"
	"burp/internal/docker"
	"burp/internal/ecosystem"
	"burp/pkg/shutdown"
	"context"
	"github.com/rs/zerolog/log"
	"time"
)

func main() {
	ecosystem.Init()
	if err := docker.Init(); err != nil {
		log.Panic().Err(err).Msg("Cannot connect to Docker")
	}
	server.Init(8873)
	<-shutdown.Shutdown(context.Background(), 5*time.Second, map[string]shutdown.Task{
		"cleanup_burp": func(ctx context.Context) error {
			return shutdown.Cleanup()
		},
	})
}
