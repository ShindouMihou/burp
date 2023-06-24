package docker

import (
	"context"
	"github.com/docker/docker/api/types"
)

func GetContainer(name string) (*types.ContainerJSON, error) {
	return returningTask(func() (types.ContainerJSON, error) {
		return Client.ContainerInspect(context.TODO(), "/"+name)
	})
}

func Kill(name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerKill(context.TODO(), name, "SIGKILL")
	})
}

func Start(name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerStart(context.TODO(), name, types.ContainerStartOptions{})
	})
}

func Remove(name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerRemove(context.TODO(), name, types.ContainerRemoveOptions{Force: true})
	})
}
