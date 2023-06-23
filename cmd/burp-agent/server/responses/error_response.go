package responses

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (errorResponse *ErrorResponse) Format(args ...string) ErrorResponse {
	text := errorResponse.Error
	for _, arg := range args {
		text = strings.Replace(text, "{$PLACEHOLDER}", arg, 1)
	}
	return ErrorResponse{Code: errorResponse.Code, Error: text}
}

func (errorResponse *ErrorResponse) Reply(ctx *gin.Context) {
	ctx.SecureJSON(errorResponse.Code, errorResponse)
}

var InvalidPayload = ErrorResponse{Code: http.StatusBadRequest, Error: "Invalid payload."}
var VagueError = ErrorResponse{Code: http.StatusBadRequest, Error: "An error occurred while trying to execute this task."}
var Unauthorized = ErrorResponse{Code: http.StatusUnauthorized, Error: "You do not have the privilege to perform this task or access this resource."}
var NotFound = ErrorResponse{Code: http.StatusNotFound, Error: "We cannot find any resource that matches."}
