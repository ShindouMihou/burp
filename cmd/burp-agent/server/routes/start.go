package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/limiter"
	"burp/cmd/burp-agent/server/requests"
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/burp"
	"burp/internal/docker"
	"burp/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

// _
// POST /application/start: You can use this route to start all the containers that are part of
// the application stack of the application.
//
// Requires: [Content-Type=multipart,File=[burp,burp.toml,application/toml]]
// Returns: sse-stream
// _
var _ = server.Add(func(app *gin.Engine) {
	app.POST("/application/start", func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Spawning server-side stream...")
		responses.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {

			responses.ChannelSend(channel, responses.Create("Waiting for deployment agent..."))
			logger.Info().Msg("Waiting for deployment agent...")

			limiter.GlobalAgentLock.Lock()
			defer limiter.GlobalAgentLock.Unlock()
			var application burp.Application
			if err := toml.Unmarshal(bytes, &application); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			for _, dependency := range application.Dependencies {
				responses.ChannelSend(channel, responses.Create("Starting dependency container (application."+dependency.Name+")...."))
				if err := docker.Start(fmt.Sprint("application.", dependency.Name)); err != nil {
					logger.Info().Err(err).Str("name", dependency.Name).Msg("Failed to start dependency container")
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to start dependency container (application."+dependency.Name+")", err.Error()))
					return
				}
				responses.ChannelSend(channel, responses.Create("Started dependency container (application."+dependency.Name+")"))
			}

			responses.ChannelSend(channel, responses.Create("Starting main service container (application."+application.Service.Name+")...."))
			if err := docker.Start(fmt.Sprint("application.", application.Service.Name)); err != nil {
				logger.Info().Err(err).Str("name", application.Service.Name).Msg("Failed to start main service container")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to start main service container (application."+application.Service.Name+")", err.Error()))
				return
			}
			responses.ChannelSend(channel, responses.Create("Started main service container (application."+application.Service.Name+")"))
			defer close(*channel)
		}()

		responses.Stream(ctx, channel)
	})
})
