package responses

import (
	"github.com/gin-gonic/gin"
	"io"
)

// AddSseHeaders is a method that adds the required headers for a server-sent stream.
func AddSseHeaders(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
}

// Stream is a method that calls Gin's stream method that will send any messages received from the
// channel down to the writer and flushes them to the client.
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
