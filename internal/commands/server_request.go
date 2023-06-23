package commands

import (
	"burp/internal/api"
	"burp/internal/commands/logins"
	"burp/pkg/console"
	"bytes"
	"github.com/AlecAivazis/survey/v2"
	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
)

var ServerRequestQuestions = []*survey.Question{
	{
		Name: "encryption",
		Prompt: &survey.Password{
			Message: "Enter an encryption key for this server. (16 characters minimum)",
			Help:    "This will be used to encrypt the credentials of the server.",
		},
		Validate: survey.MinLength(16),
	},
	{
		Name: "name",
		Prompt: &survey.Input{
			Message: "On which server do you want to perform this action? (case-insensitive)",
			Help:    "You will need this name to use it with the Burp cli.",
		},
		Validate: survey.Required,
	},
	{
		Name: "directory",
		Prompt: &survey.Input{
			Message: "Which directory from here has the Burp.toml? (relative, defaults to current)",
			Help:    "For more information, please refer to https://github.com/ShindouMihou/burp.",
		},
	},
}

type ServerRequestSurvey struct {
	api.Keys
	Directory string `json:"directory"`
}

type ServerRequestAction = func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error)

func CreateServerRequestCommand(name string, description string, action ServerRequestAction) *cli.Command {
	return &cli.Command{
		Name:        name,
		Description: description,
		Action: func(ctx *cli.Context) error {
			var answers ServerRequestSurvey
			if err := survey.Ask(ServerRequestQuestions, &answers); err != nil {
				return err
			}
			console.Clear()
			answers.Keys.Sanitize()
			secrets, ok := logins.MustUnlock(&answers.Keys)
			if !ok {
				return nil
			}
			burp, tree := api.GetBurper(answers.Directory)
			if burp == nil || tree == nil {
				return nil
			}
			request := secrets.Client().
				EnableTrace().
				SetMultipartField("burp", "burp.toml", "application/toml", bytes.NewReader(tree.Bytes())).
				SetDoNotParseResponse(true)

			api.Streamed(action(secrets, request))
			return nil
		},
	}
}
