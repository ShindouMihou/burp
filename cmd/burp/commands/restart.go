package commands

import (
	"burp/cmd/burp/api"
	"burp/cmd/burp/commands/logins"
	"burp/cmd/burp/commands/templates"
	"burp/pkg/console"
	"bytes"
	"github.com/AlecAivazis/survey/v2"
	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
)

var Restart = &cli.Command{
	Name:        "restart",
	Description: "Restarts an application on a remote server.",
	Action: func(ctx *cli.Context) error {
		var answers templates.ServerRequestSurvey
		if err := survey.Ask(templates.ServerRequestQuestions, &answers); err != nil {
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
		var createRequest = func() *resty.Request {
			return secrets.Client().
				EnableTrace().
				SetMultipartField("burp", "burp.toml", "application/toml", bytes.NewReader(tree.Bytes())).
				SetDoNotParseResponse(true)
		}
		request := createRequest()
		api.Streamed(request.Post(secrets.Link("application", "stop")))
		request = createRequest()
		api.Streamed(request.Post(secrets.Link("application", "start")))
		return nil
	},
}
