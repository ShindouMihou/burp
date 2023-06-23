package commands

import (
	"burp/cmd/burp/api"
	"github.com/go-resty/resty/v2"
)

var Stop = CreateServerRequestCommand(
	"stop",
	"Stops an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "stop"))
	},
)

var Start = CreateServerRequestCommand(
	"start",
	"Starts an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "stratrt"))
	},
)

var Remove = CreateServerRequestCommand(
	"remove",
	"Removes an application's stack on a remote server.",
	func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error) {
		return request.Post(secrets.Link("application", "remove"))
	},
)
