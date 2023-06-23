package commands

import (
	"burp/internal/api"
	"burp/internal/commands/logins"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/portainer/libcrypto"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"net/http"
	"net/url"
	"path/filepath"
)

var loginQuestions = []*survey.Question{
	{
		Name: "encryption",
		Prompt: &survey.Password{
			Message: "Enter an encryption key for this server. (Store it somewhere safe, 16 characters minimum)",
			Help:    "This will be used to encrypt the credentials of the server.",
		},
		Validate: survey.MinLength(16),
	},
	{
		Name: "server",
		Prompt: &survey.Input{
			Message: "What's the publicly-accessible link to the Burp agent?",
			Help:    "If you don't have a Burp agent installed on your server yet, please look into https://github.com/ShindouMihou/burp",
		},
		Validate: func(ans interface{}) error {
			answer := ans.(string)
			if !utils.HasPrefixStr(answer, "https://") {
				return errors.New("the link should be in HTTPS")
			}
			if _, err := url.Parse(answer); err != nil {
				return errors.Join(errors.New("invalid link format"), err)
			}
			return nil
		},
	},
	{
		Name: "name",
		Prompt: &survey.Input{
			Message: "What do you want to name this server? (case-insensitive, unique)",
			Help:    "You will need this name to use it with the Burp cli.",
		},
		Validate: survey.Required,
	},
	{
		Name: "secret",
		Prompt: &survey.Input{
			Message: "What's the Burp agent's secret token? (not hashed)",
			Help:    "For more information, please refer to https://github.com/ShindouMihou/burp.",
		},
		Validate: survey.Required,
	},
	{
		Name: "signature",
		Prompt: &survey.Input{
			Message: "What's the Burp agent's signature?",
			Help:    "For more information, please refer to https://github.com/ShindouMihou/burp.",
		},
		Validate: survey.Required,
	},
}

type loginAnswers struct {
	api.Keys
	api.Secrets
}

var Login = &cli.Command{
	Name:        "login",
	Description: "Saves and validates the credentials to a Burp agent before storing the credentials with encryption.",
	Action: func(ctx *cli.Context) error {
		var answers loginAnswers
		if err := survey.Ask(loginQuestions, &answers); err != nil {
			return err
		}
		console.Clear()
		answers.Keys.Sanitize()
		answers.Secrets.Sanitize()
		response, err := answers.Secrets.Client().Get(answers.Server)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to talk it out with Burp! It seems like something happened!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		if response.StatusCode() == http.StatusUnauthorized {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to talk it out with Burp! He said that the credentials were wrong!")
			return nil
		}
		if response.StatusCode() != http.StatusNoContent {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to talk it out with Burp! He gave us a ", response.StatusCode(), " status code!")
			return nil
		}
		bytes, err := json.Marshal(&answers.Secrets)
		if err != nil {
			return err
		}
		data, err := libcrypto.Encrypt(bytes, []byte(answers.Encryption))
		if err != nil {
			return err
		}
		file := filepath.Join(logins.Folder, answers.Name+".json")
		if err = fileutils.Save(file, data); err != nil {
			return err
		}

		fmt.Println(chalk.Green, "૮˶ᵔᵕᵔ˶ა", chalk.Reset, "Hooray, the server has been added!")
		fmt.Println(chalk.Reset, "You can use the server name ", console.Highlight, answers.Name, chalk.Reset,
			"to deploy, start, stop or remove applications with Burp.")
		return nil
	},
}
