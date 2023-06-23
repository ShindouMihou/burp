package main

import (
	commands2 "burp/cmd/burp/commands"
	"burp/cmd/burp/commands/logins"
	"burp/internal/ecosystem"
	"burp/pkg/shutdown"
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	ecosystem.Init()
	if logins.Folder == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Panic().Err(err).Msg("An error occurred.")
			return
		}
		logins.Folder = filepath.Join(home, ".burpy", "servers")
	}
	if err := commands2.App.Run(os.Args); err != nil {
		log.Panic().Err(err).Msg("An error occurred.")
	}
	shutdown.Shutdown(context.Background(), 5*time.Second, map[string]shutdown.Task{
		"cleanup_burp": func(ctx context.Context) error {
			return shutdown.Cleanup()
		},
	})
}
