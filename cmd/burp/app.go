package main

import (
	commands2 "burp/cmd/burp/commands"
	"burp/cmd/burp/commands/logins"
	"burp/internal/ecosystem"
	"burp/pkg/fileutils"
	"burp/pkg/shutdown"
	"context"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {
	ecosystem.Init()
	if logins.Folder == "" {
		logins.Folder = fileutils.JoinHomePath(".burpy", "servers")
	}
	if err := commands2.App.Run(os.Args); err != nil {
		if err == terminal.InterruptErr {
			return
		}
		log.Panic().Err(err).Msg("An error occurred.")
	}
	shutdown.Shutdown(context.Background(), 5*time.Second, map[string]shutdown.Task{
		"cleanup_burp": func(ctx context.Context) error {
			return shutdown.Cleanup()
		},
	})
}
