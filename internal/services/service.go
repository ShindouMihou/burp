package services

import (
	"burp/internal/auth"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	giturl "github.com/kubescape/go-git-url"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

var TemporaryCloneFolder = fileutils.JoinHomePath(".burpy", ".build", ".repos")

func (service *Service) Clone() (*string, error) {
	addr, err := giturl.NewGitURL(service.Repository)
	if err != nil {
		return nil, err
	}
	logger := log.With().
		Str("owner", addr.GetOwnerName()).
		Str("repository", addr.GetRepoName()).
		Str("domain", addr.GetHostName()).
		Logger()
	directory := filepath.Join(TemporaryCloneFolder, service.Name, addr.GetHostName(), addr.GetOwnerName(), addr.GetRepoName())
	if addr.GetBranchName() != "" {
		directory = filepath.Join(directory, addr.GetBranchName())
	}
	directory = strings.ToLower(directory)
	exists, err := utils.Exists(directory)
	if err != nil {
		return nil, err
	}
	if exists {
		logger.Info().Str("destination", directory).Msg("Cleaning Clone Target")
		err := os.RemoveAll(directory)
		if err != nil {
			return nil, err
		}
	}
	if err = os.MkdirAll(directory, os.ModePerm); err != nil {
		return nil, err
	}
	var transportAuth transport.AuthMethod = nil
	if creds, ok := auth.Git[addr.GetHostName()]; ok {
		log.Info().Any("creds", creds).Msg("Credential Loaded")
		transportAuth = &http.BasicAuth{
			Username: creds.Username,
			Password: creds.Password,
		}
	}
	logger.Info().Bool("authenticated", transportAuth != nil).Msg("Cloning Repository")
	_, err = git.PlainClone(directory, false, &git.CloneOptions{
		URL:  service.Repository,
		Auth: transportAuth,
	})
	if err != nil {
		return nil, err
	}
	logger.Info().Msg("Cloned")
	return &directory, nil
}

func (service *Service) GetImage() string {
	return fmt.Sprint("burp/", service.Name)
}
