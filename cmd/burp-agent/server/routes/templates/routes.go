package templates

import (
	"burp/cmd/burp-agent/server/limiter"
	"burp/cmd/burp-agent/server/requests"
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/burp"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StreamingRoute = func(ctx context.Context, channel *chan any, logger *zerolog.Logger, application *burp.Application)

func StreamingConfigOnlyRoute(action StreamingRoute) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Spawning a server-side stream...")
		responses.AddSseHeaders(ctx)
		responses.Stream(ctx, func(context context.Context, channel *chan any) {
			limiter.Await(channel, logger)
			defer limiter.GlobalAgentLock.Unlock()

			var application burp.Application
			if ok := application.From(bytes, logger, channel); !ok {
				return
			}

			action(ctx, channel, logger, &application)
		})
	}
}
