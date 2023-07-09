package main

import (
	"burp/cmd/burp/commands"
	"burp/cmd/burp/commands/logins"
	"burp/cmd/burp/commands/templates"
	"burp/internal/ecosystem"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"burp/pkg/shutdown"
	"burp/pkg/utils"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/rs/zerolog/log"
	"github.com/ttacon/chalk"
	"os"
)

func main() {
	defer shutdown.Cleanup()

	ecosystem.Init()
	if logins.Folder == "" {
		logins.Folder = fileutils.JoinHomePath(".burpy", "servers")
	}
	entries, err := os.ReadDir(logins.Folder)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Panic().Err(err).Msg("An error occurred.")
		}
		entries = []os.DirEntry{}
	}
	entries = utils.Only(entries, func(b os.DirEntry) bool {
		return b.IsDir()
	})
	logins.Servers = utils.Map(entries, func(v os.DirEntry) string {
		return v.Name()
	})
	templates.ServerRequestQuestions[0].Prompt.(*survey.Select).Options = logins.Servers
	if err := commands.App.Run(os.Args); err != nil {
		if err == terminal.InterruptErr {
			return
		}
		if err.Error() == "please provide options to select from" {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "You do not have any servers registered, please use the ", console.Highlight, "burp login", chalk.Reset, "command!")
			return
		}
		log.Panic().Err(err).Msg("An error occurred.")
	}
}
