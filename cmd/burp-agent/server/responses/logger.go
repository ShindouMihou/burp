package responses

import (
	"burp/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Logger(ctx *gin.Context) *zerolog.Logger {
	return utils.Ptr(log.With().
		Str("path", ctx.FullPath()).
		Str("method", ctx.Request.Method).
		Str("ip", ctx.ClientIP()).
		Str("user_agent", ctx.Request.UserAgent()).
		Logger())
}
