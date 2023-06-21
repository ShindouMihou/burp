package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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

func Handle(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := scanner.Bytes()
		var stream StreamResponse
		if err := json.Unmarshal(line, &stream); err != nil {
			fmt.Println("encountered an error while unmarshaling: ", string(line))
			fmt.Println("error: ", err)
			continue
		}
		if stream.Stream != nil {
			fmt.Print(*stream.Stream)
			if !strings.Contains(*stream.Stream, "\n") {
				fmt.Print("\n")
			}
		}
		if stream.Status != nil {
			fmt.Print(*stream.Status)
			if stream.ProgressDetail != nil {
				fmt.Print(" ", *stream.ProgressDetail)
			} else {
				if !strings.Contains(*stream.Status, "\n") {
					fmt.Print("\n")
				}
			}
		}
		if stream.Error != nil {
			return errors.New(*stream.Error)
		}
	}
	return nil
}
