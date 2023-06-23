package requests

import (
	"burp/cmd/burp-agent/server/mimes"
	responses2 "burp/cmd/burp-agent/server/responses"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func GetBurpFile(ctx *gin.Context) (bytes []byte, ok bool) {
	logger := responses2.Logger(ctx)

	if ctx.ContentType() != "multipart/form-data" {
		responses2.InvalidPayload.Reply(ctx)
		return nil, false
	}
	file, err := ctx.FormFile("burp")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			responses2.InvalidPayload.Reply(ctx)
			return nil, false
		}
		responses2.HandleErr(ctx, err)
		return nil, false
	}
	contentType := file.Header.Get("Content-Type")
	if contentType != mimes.TOML_MIMETYPE {
		logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
		responses2.InvalidPayload.Reply(ctx)
		return nil, false
	}
	f, err := file.Open()
	if err != nil {
		responses2.HandleErr(ctx, err)
		return nil, false
	}
	bytes, err = io.ReadAll(f)
	if err != nil {
		responses2.HandleErr(ctx, err)
		return nil, false
	}
	return bytes, true
}
