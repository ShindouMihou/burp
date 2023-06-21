package docker

import (
	"burp/utils"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var Client *client.Client

func Init() error {
	c, err := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	Client = c
	return nil
}

func HasContainer(name string) (bool, error) {
	containers, err := Client.ContainerList(context.TODO(), types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return false, err
	}
	for _, container := range containers {
		if utils.AnyMatchString(container.Names, name) {
			return true, nil
		}
	}
	return false, nil
}
