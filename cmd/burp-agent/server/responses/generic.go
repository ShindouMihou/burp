package responses

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Arrayed struct {
	Data any `json:"data"`
}

type GenericResponse struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// CreateChannelError creates an ErrorResponse that is more suited towards server-sent events which needs to be
// more verbose than usual since the responses tend to be printed to the console.
func CreateChannelError(context string, error string) *ErrorResponse {
	return &ErrorResponse{Error: fmt.Sprint(context, ": ", error), Code: http.StatusBadGateway}
}

// Create creates a GenericResponse that tends to indicate that something is happening or the stream is still processing
// and is doing alright, this can also be used to return back general data back to any request.
func Create(data any) *GenericResponse {
	return &GenericResponse{
		Code: http.StatusOK,
		Data: data,
	}
}

// Error is a shorthand method that creates an ErrorResponse using the given parameters and sends it down to the
// channel that is provided.
func Error(channel *chan any, context string, err error) {
	ChannelSend(channel, &ErrorResponse{Error: fmt.Sprint(context, ": ", err.Error()), Code: http.StatusBadGateway})
}

// Message is a shorthand method that joins the messages using Sprint before passing it down to
// Create which creates a  generic response that is then passed to ChannelSend.
func Message(channel *chan any, msg ...any) {
	ChannelSend(channel, Create(fmt.Sprint(msg...)))
}

// ChannelSend is a longer way of doing "channel <- data". All server-sent events should use this method to send to
// their channel as it will allow us to perform other side effects in the far future when needed, such as, closing
// the channel when it receives an error response.
func ChannelSend(channel *chan any, data any) {
	if channel != nil {
		*channel <- data
	}
}
