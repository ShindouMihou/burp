package services

import (
	"burp/utils"
	"fmt"
	"github.com/go-git/go-git/v5"
	giturl "github.com/kubescape/go-git-url"
	"os"
	"strings"
)

func (service *Service) Clone() (*string, error) {
	addr, err := giturl.NewGitURL(service.Repository)
	if err != nil {
		return nil, err
	}
	directory := fmt.Sprint(".burp/.temp/", addr.GetHostName(), "/", addr.GetOwnerName(), "/", addr.GetRepoName(), "/")
	if addr.GetBranchName() != "" {
		directory += addr.GetBranchName() + "/"
	}
	directory = strings.ToLower(directory)
	exists, err := utils.Exists(directory)
	if err != nil {
		return nil, err
	}
	if exists {
		err := os.RemoveAll(directory)
		if err != nil {
			return nil, err
		}
	}
	if err = os.MkdirAll(directory, os.ModePerm); err != nil {
		return nil, err
	}
	_, err = git.PlainClone(directory, false, &git.CloneOptions{
		URL: service.Repository,
	})
	if err != nil {
		return nil, err
	}
	return &directory, nil
}

func (service *Service) GetImage() string {
	return fmt.Sprint("burp/", service.Name)
}
