package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/routes/templates"
	"burp/internal/burp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// _
// POST /application/start: You can use this route to start all the containers that are part of
// the application stack of the application.
//
// Requires: [Content-Type=multipart,File=[burp,burp.toml,application/toml]]
// Returns: sse-stream
// _
var _ = server.Add(func(app *gin.Engine) {
	app.POST("/application/start", templates.StreamingConfigOnlyRoute(
		func(channel *chan any, logger *zerolog.Logger, application *burp.Application) {
			if ok := application.Service.Start(channel, logger); !ok {
				return
			}
			for _, dependency := range application.Dependencies {
				if ok := dependency.Start(channel, logger); !ok {
					return
				}
			}
		},
	))
})
