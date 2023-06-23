package main

import (
	commands2 "burp/internal/commands"
	"burp/internal/commands/logins"
	"burp/internal/ecosystem"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
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
}
