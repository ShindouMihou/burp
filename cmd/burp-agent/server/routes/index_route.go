package routes

import (
	"burp/cmd/burp-agent/server"
	"github.com/gin-gonic/gin"
	"net/http"
)

// _
// GET /: You can use this route to test the authentication of your client.
// Returns: [status=204, body=]
// _
var _ = server.Add(func(app *gin.Engine) {
	app.GET("/", func(context *gin.Context) { context.Status(http.StatusNoContent) })
})
