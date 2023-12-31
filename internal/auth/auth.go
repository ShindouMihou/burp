package auth

import (
	"bufio"
	"burp/pkg/env"
	"burp/pkg/fileutils"
	"bytes"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

type Authentication struct {
	Domain   string `toml:"domain" json:"domain"`
	Username string `toml:"username" json:"-"`
	Password string `toml:"password" json:"-"`
}

var Git = make(map[string]Authentication)
var Docker = make(map[string]Authentication)

type AuthenticationToml struct {
	Auth []Authentication `toml:"auth" json:"auth"`
}

func Add(file string, store map[string]Authentication) error {
	auth := AuthenticationToml{}
	f, err := fileutils.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	scanner := bufio.NewScanner(f)
	var b bytes.Buffer
	for scanner.Scan() {
		b.Write(scanner.Bytes())
	}
	err = toml.Unmarshal(b.Bytes(), &auth)
	if err != nil {
		return err
	}
	for _, creds := range auth.Auth {
		creds := creds
		log.Info().Any("creds", creds).Msg("Credential Loaded")
		store[strings.ToLower(creds.Domain)] = creds
	}
	return nil
}

func Load() {
	err := Add(env.GitToml.Or(fileutils.JoinHomePath(".burpy", ".secrets", "git.toml")), Git)
	if err != nil {
		log.Err(err)
		return
	}
	err = Add(env.DockerToml.Or(fileutils.JoinHomePath(".burpy", ".secrets", "docker.toml")), Docker)
	if err != nil {
		log.Err(err)
		return
	}
}
