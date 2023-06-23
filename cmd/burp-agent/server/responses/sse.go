package responses

import (
	"github.com/gin-gonic/gin"
	"io"
)

func AddSseHeaders(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
}

func Stream(ctx *gin.Context, channel *chan any) {
	ctx.Stream(func(w io.Writer) bool {
		if msg, ok := <-*channel; ok {
			Logger(ctx).Info().Any("data", msg).Msg("Sent stream message")
			ctx.SSEvent("data", msg)
			return true
		}
		return false
	})
}
