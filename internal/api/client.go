package api

import (
	"bufio"
	"burp/internal/server/responses"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/ttacon/chalk"
	"io"
	"net/http"
	"net/url"
)

var Client = resty.New().
	SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

func Create(secrets *Secrets) *resty.Request {
	return Client.R().
		SetHeader("X-Burp-Signature", secrets.Signature).
		SetAuthToken(secrets.Secret)
}

type Keys struct {
	Encryption string `json:"-"`
	Name       string `json:"name"`
}

func (keys *Keys) Sanitize() {
	keys.Name = fileutils.Sanitize(keys.Name)
}

type Secrets struct {
	Server    string `json:"server"`
	Secret    string `json:"secret"`
	Signature string `json:"signature"`
}

func (secrets *Secrets) Link(paths ...string) string {
	u, _ := url.JoinPath(secrets.Server, paths...)
	return u
}

func (secrets *Secrets) Client() *resty.Request {
	return Create(secrets)
}

func (secrets *Secrets) Sanitize() {
	uri, _ := url.Parse(secrets.Server)
	secrets.Server = "https://" + uri.Hostname()
	if uri.Port() != "" {
		secrets.Server += ":" + uri.Port()
	}
}

var ErrorKeyword = []byte{'"', 'e', 'r', 'r', 'o', 'r', '"', ':'}
var DataKeyword = []byte{'"', 'd', 'a', 't', 'a', '"', ':'}
var SseDataKeyword = []byte{'d', 'a', 't', 'a', ':'}

func Streamed(response *resty.Response, err error) {
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to get a message from the server!")
		fmt.Println(chalk.Red, err.Error())
		return
	}

	if response.StatusCode() == http.StatusUnauthorized {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to talk it out with Burp! He said that the credentials were wrong!")
		return
	}

	if response.StatusCode() != http.StatusOK {
		body := response.Body()
		if len(body) != 0 {
			var errorResponse responses.ErrorResponse
			if err := json.Unmarshal(body, &errorResponse); err != nil {
				fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to decode the secret error message!")
				fmt.Println(chalk.Red, err.Error())
				return
			}
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We got a secret error message from the nuclear reactor!")
			fmt.Println(chalk.Red, errorResponse.Error)
			return
		} else {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to talk it out with Burp! He gave us a ", response.StatusCode(), " status code!")
		}
		return
	}
	body := response.RawBody()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Failed to close body!")
		}
	}(body)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Bytes()

		if !utils.HasPrefix(line, SseDataKeyword) {
			continue
		}
		line = line[len(SseDataKeyword):]
		// IMPT: If the server sends back an error response, the server would've also canceled
		// the connection between us.
		if bytes.Contains(line, ErrorKeyword) {
			var errorResponse responses.ErrorResponse
			if err := json.Unmarshal(line, &errorResponse); err != nil {
				fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to decode the secret error message!")
				fmt.Println(chalk.Red, err.Error())
				return
			}
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We got a secret error message from the nuclear reactor!")
			fmt.Println(chalk.Red, errorResponse.Error)
			return
		}

		if bytes.Contains(line, DataKeyword) {
			var genericResponse responses.GenericResponse
			if err := json.Unmarshal(line, &genericResponse); err != nil {
				fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to decode the secret message!")
				fmt.Println(chalk.Red, err.Error())
				continue
			}
			fmt.Println(chalk.Yellow, "AGENT: ", genericResponse.Data)
		}
	}
}
