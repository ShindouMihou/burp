package commands

import (
	"burp/cmd/burp/api"
	"burp/cmd/burp/commands/templates"
	"github.com/go-resty/resty/v2"
)

var Stop = templates.CreateServerRequestCommand(
	"stop",
	"Stops an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "stop"))
	},
)

var Start = templates.CreateServerRequestCommand(
	"start",
	"Starts an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "stratrt"))
	},
)

var Remove = templates.CreateServerRequestCommand(
	"remove",
	"Removes an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "remove"))
	},
)
