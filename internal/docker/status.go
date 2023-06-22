package docker

import (
	"bufio"
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
	Status         *string `json:"status"`
	ProgressDetail *string `json:"progressDetail"`
	Progress       *string `json:"progress"`
	Id             *string `json:"id"`
}

type ErrorLine struct {
	Error       *string      `json:"error"`
	ErrorDetail *ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func Handle(log zerolog.Logger, scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := scanner.Bytes()
		var stream StreamResponse
		if err := json.Unmarshal(line, &stream); err != nil {
			log.Err(err).Str("line", string(line)).Msg("Failed Unmarshal")
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
				log.Print(str)
			}
		}
		if stream.Status != nil {
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
				log.Print(str)
			}
			if stream.ProgressDetail != nil {
				log.Print(" ", *stream.ProgressDetail)
			}
		}
		if stream.Error != nil {
			return errors.New(*stream.Error)
		}
	}
	return nil
}
