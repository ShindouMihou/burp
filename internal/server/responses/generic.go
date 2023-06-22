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

func CreateChannelError(context string, error string) *ErrorResponse {
	return &ErrorResponse{Error: fmt.Sprint(context, ": ", error), Code: http.StatusBadGateway}
}

func CreateChannelOk(data any) *GenericResponse {
	return &GenericResponse{
		Code: http.StatusOK,
		Data: data,
	}
}

func ChannelSend(channel *chan any, data any) {
	if channel != nil {
		*channel <- data
		if _, ok := data.(*ErrorResponse); ok {
			close(*channel)
		}
	}
}
