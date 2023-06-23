package requests

import (
	"burp/cmd/burp-agent/server/mimes"
	responses "burp/cmd/burp-agent/server/responses"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

// GetBurpFile is a utility method to get the Burp file from a single-file upload route,
// this checks the mimetype and reads all the bytes before returning it.
func GetBurpFile(ctx *gin.Context) (bytes []byte, ok bool) {
	logger := responses.Logger(ctx)

	if ctx.ContentType() != "multipart/form-data" {
		responses.InvalidPayload.Reply(ctx)
		return nil, false
	}
	file, err := ctx.FormFile("burp")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			responses.InvalidPayload.Reply(ctx)
			return nil, false
		}
		responses.HandleErr(ctx, err)
		return nil, false
	}
	contentType := file.Header.Get("Content-Type")
	if contentType != mimes.TOML_MIMETYPE {
		logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
		responses.InvalidPayload.Reply(ctx)
		return nil, false
	}
	f, err := file.Open()
	if err != nil {
		responses.HandleErr(ctx, err)
		return nil, false
	}
	bytes, err = io.ReadAll(f)
	if err != nil {
		responses.HandleErr(ctx, err)
		return nil, false
	}
	return bytes, true
}
