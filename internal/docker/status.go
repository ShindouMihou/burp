package docker

import (
	"bufio"
	"burp/cmd/burp-agent/server/responses"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog"
	"strings"
)

type StreamResponse struct {
	Stream *string `json:"stream"`
	Aux    *Aux    `json:"aux"`
	ErrorLine
	Progressing
}

type Aux struct {
	Id string `json:"ID"`
}

type Progressing struct {
	Status   *string `json:"status,omitempty"`
	Progress *string `json:"progress,omitempty"`
	Id       *string `json:"id,omitempty"`
}

type ErrorLine struct {
	Error       *string      `json:"error,omitempty"`
	ErrorDetail *ErrorDetail `json:"errorDetail,omitempty"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func Handle(channel *chan any, log zerolog.Logger, scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := scanner.Bytes()
		var stream StreamResponse
		if err := json.Unmarshal(line, &stream); err != nil {
			log.Err(err).Str("line", string(line)).Msg("Failed Unmarshal")
			responses.Error(channel, "Failed to unmarshal status", err)
			continue
		}
		if stream.Stream != nil {
			streams := strings.Split(*stream.Stream, "\n")
			for _, str := range streams {
				str := strings.ReplaceAll(str, "\n", "")
				str = strings.TrimSpace(str)
				if len(str) == 0 {
					continue
				}
				if strings.HasPrefix(str, "\u001b[0m") {
					continue
				}
				responses.Message(channel, str)
				log.Print(str)
			}
		}
		if stream.Status != nil {
			if strings.Contains(*stream.Status, "\n") {
				streams := strings.Split(*stream.Status, "\n")
				for _, str := range streams {
					str := strings.ReplaceAll(str, "\n", "")
					str = strings.TrimSpace(str)
					if len(str) == 0 {
						continue
					}
					if strings.HasPrefix(str, "\u001b[0m") {
						continue
					}
					responses.Message(channel, str)
					log.Print(str)
				}
				if stream.Progress != nil {
					log.Print(" ", *stream.Progress)
				}
			} else {
				log.Print(*stream.Status)
				if stream.Progress != nil {
					log.Print(" ", *stream.Progress)
					responses.Message(channel, *stream.Status, " ", *stream.Progress)
				} else {
					responses.Message(channel, *stream.Status)
				}
			}

		}
		if stream.Error != nil {
			responses.ChannelSend(channel, responses.CreateChannelError("An error occurred", *stream.Error))
			return errors.New(*stream.Error)
		}
	}
	return nil
}
