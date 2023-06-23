package main

import (
	"burp/cmd/burp/commands"
	"burp/cmd/burp/commands/logins"
	"burp/internal/ecosystem"
	"burp/pkg/fileutils"
	"burp/pkg/shutdown"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	defer shutdown.Cleanup()

	ecosystem.Init()
	if logins.Folder == "" {
		logins.Folder = fileutils.JoinHomePath(".burpy", "servers")
	}
	if err := commands.App.Run(os.Args); err != nil {
		if err == terminal.InterruptErr {
			return
		}
		log.Panic().Err(err).Msg("An error occurred.")
	}
}
