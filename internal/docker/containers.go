package docker

import (
	"context"
	"github.com/docker/docker/api/types"
)

func GetContainer(name string) (*types.ContainerJSON, error) {
	return returningTask(func() (types.ContainerJSON, error) {
		return Client.ContainerInspect(context.Background(), "/"+name)
	})
}

func Kill(ctx context.Context, name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerKill(ctx, name, "SIGKILL")
	})
}

func Start(ctx context.Context, name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerStart(ctx, name, types.ContainerStartOptions{})
	})
}

func Remove(ctx context.Context, name string) error {
	return nonReturningTask(func() error {
		return Client.ContainerRemove(ctx, name, types.ContainerRemoveOptions{Force: true})
	})
}
