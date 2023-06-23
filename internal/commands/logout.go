package commands

import (
	"burp/internal/commands/logins"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var logoutConfirmation = []*survey.Question{
	{
		Name: "confirmation",
		Prompt: &survey.Confirm{
			Message: "Are you sure you want to delete this server? This is an irreversible action!",
			Default: false,
		},
		Validate: survey.Required,
	},
}

var Logout = &cli.Command{
	Name:        "logout",
	Description: "Removes a server from the saved logins.",
	Action: func(ctx *cli.Context) error {
		name := ctx.Args().First()
		if name == "" {
			fmt.Println(chalk.Red, "૮₍˃⤙˂₎ა", chalk.Reset, "You need to supply the ", console.Highlight, "server name", chalk.Reset, "to log out from!")
			return nil
		}
		var confirmation Confirmation
		if err := survey.Ask(logoutConfirmation, &confirmation); err != nil {
			return err
		}
		console.Clear()
		if !confirmation.Confirmation {
			return nil
		}
		name = fileutils.Sanitize(name)
		file := filepath.Join(logins.Folder, name+".json")
		if err := os.Remove(file); err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to negotiate with the operating system!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		fmt.Println(chalk.Green, "૮˶ᵔᵕᵔ˶ა", chalk.Reset, "Hooray, you are now logged out from the server!")
		return nil
	},
}
