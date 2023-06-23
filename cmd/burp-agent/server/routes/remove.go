package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/limiter"
	"burp/cmd/burp-agent/server/requests"
	responses "burp/cmd/burp-agent/server/responses"
	"burp/internal/docker"
	"burp/internal/services"
	"burp/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

// _
// POST /application/remove: You can use this route to remove all the containers that are part of
// the application stack of the application.
//
// Requires: [Content-Type=multipart,File=[burp,burp.toml,application/toml]]
// Returns: sse-stream
// _
var _ = server.Add(func(app *gin.Engine) {
	app.POST("/application/remove", func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Spawning a server-side stream...")
		responses.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {
			responses.ChannelSend(channel, responses.Create("Waiting for deployment agent..."))
			logger.Info().Msg("Waiting for deployment agent...")

			limiter.GlobalAgentLock.Lock()
			defer limiter.GlobalAgentLock.Unlock()

			var burp services.Burp
			if err := toml.Unmarshal(bytes, &burp); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			responses.ChannelSend(channel, responses.Create("Removing main service container (burp."+burp.Service.Name+")...."))
			if err := docker.Remove(fmt.Sprint("burp.", burp.Service.Name)); err != nil {
				logger.Info().Err(err).Str("name", burp.Service.Name).Msg("Failed to remove main service container")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to remove service container (burp."+burp.Service.Name+")", err.Error()))
				return
			}
			responses.ChannelSend(channel, responses.Create("Removed main service container (burp."+burp.Service.Name+")"))
			for _, dependency := range burp.Dependencies {
				responses.ChannelSend(channel, responses.Create("Removing dependency container (burp."+dependency.Name+")...."))
				if err := docker.Remove(fmt.Sprint("burp.", dependency.Name)); err != nil {
					logger.Info().Err(err).Str("name", dependency.Name).Msg("Failed to remove dependency container")
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to remove dependency container (burp."+dependency.Name+")", err.Error()))
					return
				}
				responses.ChannelSend(channel, responses.Create("Removed dependency container (burp."+dependency.Name+")"))
			}
			defer close(*channel)
		}()

		responses.Stream(ctx, channel)
	})
})
