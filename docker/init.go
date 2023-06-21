package docker

import (
	"github.com/docker/docker/client"
)

var Client *client.Client

func Init() error {
	c, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	Client = c
	return nil
}
