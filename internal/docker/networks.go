package docker

import (
	"context"
	"github.com/docker/docker/api/types"
)

func HasNetwork(name string) (bool, error) {
	return has(func() (any, error) {
		return Client.NetworkInspect(context.TODO(), name, types.NetworkInspectOptions{})
	})
}
