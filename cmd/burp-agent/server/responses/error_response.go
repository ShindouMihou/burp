package responses

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// Format formats an ErrorResponse, replacing whatever {$PLACEHOLDER} out there with
// the given values. This may be used right now, and maybe has no place, but it'll exist.
func (errorResponse *ErrorResponse) Format(args ...string) ErrorResponse {
	text := errorResponse.Error
	for _, arg := range args {
		text = strings.Replace(text, "{$PLACEHOLDER}", arg, 1)
	}
	return ErrorResponse{Code: errorResponse.Code, Error: text}
}

// Reply is a short-hand method that allows for an idiomatic way of replying to the
// request with an error response.
func (errorResponse *ErrorResponse) Reply(ctx *gin.Context) {
	ctx.SecureJSON(errorResponse.Code, errorResponse)
}

// HandleErr is a utility method that handles errors on a request, this sends a VagueError to the client
// and logs the error into the console.
func HandleErr(ctx *gin.Context, err error) {
	VagueError.Reply(ctx)
	Logger(ctx).Err(err).Msg("Encountered an Error")
}

var InvalidPayload = ErrorResponse{Code: http.StatusBadRequest, Error: "Invalid payload."}
var VagueError = ErrorResponse{Code: http.StatusBadRequest, Error: "An error occurred while trying to execute this task."}
var Unauthorized = ErrorResponse{Code: http.StatusUnauthorized, Error: "You do not have the privilege to perform this task or access this resource."}
var NotFound = ErrorResponse{Code: http.StatusNotFound, Error: "We cannot find any resource that matches."}
