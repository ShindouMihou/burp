package api

import (
	"bufio"
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/responses"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/ttacon/chalk"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
)

var InsecureClient = resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

func CreateClientWithTls(name string, secrets *Secrets) (*resty.Request, error) {
	pool, err := CreateCertificatePool(name)
	if err != nil {
		return nil, err
	}
	client := resty.New().SetTLSClientConfig(&tls.Config{RootCAs: pool})
	return CreateWithClient(client, secrets), nil
}

func CreateCertificatePool(name string) (*x509.CertPool, error) {
	rootCertificates, _ := x509.SystemCertPool()
	if rootCertificates == nil {
		rootCertificates = x509.NewCertPool()
	}

	sslFileName := filepath.Join(server.TemporarySslDirectory, name, "ssl.cert")
	file, err := fileutils.Open(sslFileName)
	if err != nil {
		return nil, err
	}
	certificate, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	rootCertificates.AppendCertsFromPEM(certificate)
	return rootCertificates, nil
}

func CreateWithClient(client *resty.Client, secrets *Secrets) *resty.Request {
	return client.R().
		SetHeader("X-Burp-Signature", secrets.Signature).
		SetAuthToken(secrets.Secret)
}

func CreateInsecure() *resty.Request {
	return InsecureClient.R()
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

func (secrets *Secrets) ClientWithTls(name string) (*resty.Request, bool) {
	client, err := CreateClientWithTls(name, secrets)
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We cannot verify the server's authenticity!")
		fmt.Println(chalk.Red, "It is likely that the SSL certificate cannot be found, try removing then adding this server again.")
		fmt.Println(chalk.Red, err.Error())
		return nil, false
	}
	return client, true
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
