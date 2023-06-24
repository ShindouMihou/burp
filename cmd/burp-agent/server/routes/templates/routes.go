package templates

import (
	"burp/cmd/burp-agent/server/limiter"
	"burp/cmd/burp-agent/server/requests"
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/burp"
	"burp/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StreamingRoute = func(channel *chan any, logger *zerolog.Logger, application *burp.Application)

func StreamingConfigOnlyRoute(action StreamingRoute) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Spawning a server-side stream...")
		responses.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {
			limiter.Await(channel, logger)
			defer limiter.GlobalAgentLock.Unlock()

			var application burp.Application
			if ok := application.From(bytes, logger, channel); !ok {
				return
			}

			defer close(*channel)
			action(channel, logger, &application)
		}()

		responses.Stream(ctx, channel)
	}
}
