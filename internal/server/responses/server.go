package responses

import (
	"github.com/gin-gonic/gin"
)

func HandleErr(ctx *gin.Context, err error) {
	VagueError.Reply(ctx)
	Logger(ctx).Err(err).Msg("Encountered an Error")
}
